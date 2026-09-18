// 桥接服务（全栈开发1）
//
// 作用：原后端 main.go 是一个「终端交互式」控制台程序（fmt.Scan 读入，
// fmt.Println 输出，用户数据保存在内存切片 userlist 中）。后端代码不可更改，
// 所以这里不重写任何一句后端逻辑，而是启动它、用管道把网页请求喂给它的
// 标准输入，并解析它打印出来的中文提示，再以 HTTP JSON 的形式返回给浏览器。
//
// 数据流：
//   浏览器(web/) --HTTP JSON--> 桥接服务 --stdin/stdout 管道--> main.go（原样运行）
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ---------------------------------------------------------------- 后端进程管理

type token struct {
	text string
	err  error
}

// backend 持有那个「不可修改」的控制台程序，并在整个网页会话期间让它常驻，
// 这样 userlist（内存中的已注册用户）就不会因为进程重启而丢失。
type backend struct {
	mu       sync.Mutex // 串行化一次操作，避免多个请求把终端输入交错
	srcPath  string     // 原始后端源码 main.go
	exePath  string     // 编译产物（源码本身没有被修改过）
	workDir  string
	cmd      *exec.Cmd
	stdin    io.WriteCloser
	tokens   chan token
	alive    bool
	exitCode string
}

func newBackend(srcPath, workDir string) *backend {
	return &backend{srcPath: srcPath, workDir: workDir}
}

// build 编译原后端源码（只读，不修改文件内容）。
func (b *backend) build() error {
	binDir := filepath.Join(b.workDir, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return fmt.Errorf("创建 bin 目录失败：%w", err)
	}
	exe := filepath.Join(binDir, "loginbackend.exe")

	build := exec.Command("go", "build", "-o", exe, b.srcPath)
	build.Dir = filepath.Dir(b.srcPath)
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		return fmt.Errorf("编译后端 %s 失败（请确认已安装 Go 并加入 PATH）：%w", b.srcPath, err)
	}
	b.exePath = exe
	return nil
}

// start 启动常驻后端进程，并等待它打印出主菜单。
func (b *backend) start() error {
	if b.alive {
		return nil
	}
	if b.exePath == "" {
		if err := b.build(); err != nil {
			return err
		}
	}

	cmd := exec.Command(b.exePath)
	cmd.Dir = filepath.Dir(b.srcPath) // 与用户手动运行 const 程序时的工作目录保持一致

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动后端进程失败：%w", err)
	}

	b.cmd = cmd
	b.stdin = stdin
	b.tokens = make(chan token, 256)
	b.alive = true
	b.exitCode = ""

	go b.readLoop(stdout)
	go func() {
		err := cmd.Wait()
		code := "0"
		if err != nil {
			code = err.Error()
		}
		b.mu.Lock()
		if b.cmd == cmd {
			b.alive = false
			b.exitCode = code
		}
		b.mu.Unlock()
	}()

	// 后端启动后立刻打印主菜单，读到菜单结尾才算真正就绪。
	if _, err := b.awaitMenu(8 * time.Second); err != nil {
		return fmt.Errorf("后端已启动但没有输出主菜单：%w", err)
	}
	return nil
}

// readLoop 把后端的输出切成「token」。
//
// 后端有两种输出：结果提示（Println，以 \n 结尾），以及不带换行的输入提示
// 「输入你的用户名>」。因此这里同时以 '>' 与 '\n' 作为分隔符，
// 保证提示语一到就能立刻回应，不会卡住。
func (b *backend) readLoop(r io.Reader) {
	br := bufio.NewReader(r)
	var sb strings.Builder
	for {
		c, err := br.ReadByte()
		if err != nil {
			if sb.Len() > 0 {
				b.tokens <- token{text: sb.String()}
			}
			b.tokens <- token{err: err}
			return
		}
		sb.WriteByte(c)
		if c == '>' || c == '\n' {
			b.tokens <- token{text: sb.String()}
			sb.Reset()
		}
	}
}

var errTimeout = errors.New("等待后端输出超时")

func (b *backend) next(d time.Duration) (string, error) {
	select {
	case t := <-b.tokens:
		if t.err != nil {
			if errors.Is(t.err, io.EOF) {
				return "", io.EOF
			}
			return "", t.err
		}
		return t.text, nil
	case <-time.After(d):
		return "", errTimeout
	}
}

func (b *backend) write(s string) error {
	if !b.alive {
		return errors.New("后端进程未运行")
	}
	if _, err := io.WriteString(b.stdin, s); err != nil {
		return fmt.Errorf("写入后端进程失败：%w", err)
	}
	return nil
}

// awaitMenu 一直读到主菜单的最后一行，作为「后端空闲、可以接受新命令」的同步点。
func (b *backend) awaitMenu(d time.Duration) ([]string, error) {
	var lines []string
	deadline := time.Now().Add(d)
	for {
		left := time.Until(deadline)
		if left <= 0 {
			return lines, errTimeout
		}
		tok, err := b.next(left)
		if err != nil {
			return lines, err
		}
		lines = append(lines, tok)
		if strings.Contains(tok, "退出程序") {
			return lines, nil
		}
	}
}

// operation 执行一次用户操作：1 注册 / 2 登录 / 3 退出。
func (b *backend) operation(action int, username, password string) (*Result, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if err := b.start(); err != nil {
		return nil, err
	}

	res := &Result{Action: action, ActionName: actionName(action), Output: []string{}}
	record := func(tok string) {
		if s := strings.TrimRight(tok, "\n"); s != "" {
			res.Output = append(res.Output, s)
		}
	}

	if err := b.write(fmt.Sprintf("%d\n", action)); err != nil {
		return nil, err
	}
	log.Printf("→ 发送给后端：%d（%s）", action, res.ActionName)

	const stepTimeout = 10 * time.Second

	for {
		tok, err := b.next(stepTimeout)
		if err != nil {
			if errors.Is(err, io.EOF) {
				res.Code, res.OK, res.Message = "backend_exited", false, "后端程序已退出"
				return res, nil
			}
			return nil, err
		}
		record(tok)

		switch {
		// 用户名提示 → 送入用户名
		case strings.HasSuffix(tok, ">") && strings.Contains(tok, "用户名"):
			if err := b.write(username + "\n"); err != nil {
				return nil, err
			}
			log.Printf("→ 发送给后端：用户名 %q", username)
		// 密码提示 → 送入密码（后端自行做 SHA-256 哈希，桥接层不接触哈希）
		case strings.HasSuffix(tok, ">") && strings.Contains(tok, "密码"):
			if err := b.write(password + "\n"); err != nil {
				return nil, err
			}
			log.Printf("→ 发送给后端：密码（%d 个字符，仅转发）", len(password))

		// 后端的结果提示，逐字对应 main.go 里 Println 的内容
		case strings.Contains(tok, "注册成功"):
			res.Code, res.OK, res.Message = "register_success", true, "注册成功"
		case strings.Contains(tok, "该用户名已被注册"):
			res.Code, res.OK, res.Message = "duplicate_username", false, "该用户名已被注册"
		case strings.Contains(tok, "登录成功"):
			res.Code, res.OK, res.Message = "login_success", true, "登录成功"
		case strings.Contains(tok, "密码错误"):
			res.Code, res.OK, res.Message = "wrong_password", false, "密码错误"
		case strings.Contains(tok, "用户不存在"):
			res.Code, res.OK, res.Message = "user_not_found", false, "用户不存在"
		case strings.Contains(tok, "退出登录"):
			res.Code, res.OK, res.Message = "exit_ok", true, "已退出后端程序"
		case strings.Contains(tok, "错误命令"):
			res.Code, res.OK, res.Message = "invalid_command", false, "错误命令"
		}

		if res.Code == "" {
			continue
		}

		if action == 3 {
			b.shutdown() // 后端 os.Exit(0)，这里回收进程状态
			break
		}
		// 其余操作：继续读到主菜单结尾，保证下一次操作的输入不会错位
		if _, err := b.awaitMenu(3 * time.Second); err != nil {
			res.Output = append(res.Output, "（提示：主菜单回读超时，已忽略）")
		}
		break
	}

	return res, nil
}

// shutdown 等待退出的后端进程结束，并复位状态（下次操作会自动重新启动）。
func (b *backend) shutdown() {
	if b.stdin != nil {
		b.stdin.Close()
	}
	b.alive = false
	b.exitCode = "0"
	log.Printf("后端程序已退出（exit 0）")
}

// ---------------------------------------------------------------- HTTP 接口

type Result struct {
	Action     int      `json:"action"`
	ActionName string   `json:"actionName"`
	Code       string   `json:"code"`
	OK         bool     `json:"ok"`
	Message    string   `json:"message"`
	Output     []string `json:"output"` // 后端原始终端输出，便于在页面上核对
}

type errorBody struct {
	Code    string `json:"code"`
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

type credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func actionName(a int) string {
	switch a {
	case 1:
		return "注册"
	case 2:
		return "登录"
	case 3:
		return "退出"
	}
	return "未知"
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// 允许直接用 file:// 打开 index.html 时也能访问桥接服务（课堂演示用）
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func decodeCredentials(r *http.Request) (credentials, error) {
	var c credentials
	if err := json.NewDecoder(io.LimitReader(r.Body, 4096)).Decode(&c); err != nil {
		return c, errors.New("请求内容不是合法的 JSON")
	}
	c.Username = strings.TrimSpace(c.Username)
	if c.Username == "" {
		return c, errors.New("用户名不能为空")
	}
	if c.Password == "" {
		return c, errors.New("密码不能为空")
	}
	if len([]rune(c.Username)) > 32 {
		return c, errors.New("用户名过长（最多 32 个字符）")
	}
	if len([]rune(c.Password)) > 64 {
		return c, errors.New("密码过长（最多 64 个字符）")
	}
	if strings.ContainsAny(c.Username, " \t\r\n") || strings.ContainsAny(c.Password, " \t\r\n") {
		// 后端用 fmt.Scan 按空白分词读取，带空格会被拆成两个词，这里提前拦住
		return c, errors.New("用户名和密码里不能包含空格或换行")
	}
	return c, nil
}

func handleOperation(b *backend, action int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, errorBody{"method_not_allowed", false, "请使用 POST 请求"})
			return
		}
		var (
			res *Result
			err error
		)
		if action == 3 {
			res, err = b.operation(3, "", "")
		} else {
			var c credentials
			c, err = decodeCredentials(r)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, errorBody{"invalid_input", false, err.Error()})
				return
			}
			res, err = b.operation(action, c.Username, c.Password)
		}
		if err != nil {
			log.Printf("操作失败：%v", err)
			writeJSON(w, http.StatusServiceUnavailable, errorBody{"backend_unavailable", false, err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, res)
	}
}

func handleStatus(b *backend) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b.mu.Lock()
		alive, code := b.alive, b.exitCode
		b.mu.Unlock()
		state := "已退出"
		if alive || code == "" {
			state = "运行中（内存中保留了已注册用户）"
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":     true,
			"alive":  alive,
			"state":  state,
			"source": b.srcPath,
		})
	}
}

// ---------------------------------------------------------------- 启动

func firstExisting(paths ...string) string {
	for _, p := range paths {
		if p == "" {
			continue
		}
		if abs, err := filepath.Abs(p); err == nil {
			if st, err := os.Stat(abs); err == nil && !st.IsDir() {
				return abs
			}
		}
	}
	return ""
}

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "网页与接口监听的地址")
	backendPath := flag.String("backend", "../main.go", "后端源码路径（原样运行，不修改）")
	webDir := flag.String("web", "../web", "前端静态文件目录")
	open := flag.Bool("open", true, "启动后自动打开浏览器")
	flag.Parse()

	workDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("无法获取工作目录：%v", err)
	}
	exeDir := filepath.Dir(os.Args[0])

	src := firstExisting(
		*backendPath,
		filepath.Join("..", "main.go"),
		filepath.Join(workDir, "..", "main.go"),
		filepath.Join(exeDir, "..", "main.go"),
		filepath.Join(exeDir, "main.go"),
	)
	if src == "" {
		log.Fatalf("找不到后端源码 main.go（可用 -backend 指定路径）")
	}

	web := ""
	for _, cand := range []string{*webDir, filepath.Join("..", "web"), filepath.Join(workDir, "..", "web"), filepath.Join(exeDir, "..", "web")} {
		if cand == "" {
			continue
		}
		if abs, err := filepath.Abs(cand); err == nil {
			if st, err := os.Stat(abs); err == nil && st.IsDir() {
				web = abs
				break
			}
		}
	}
	if web == "" {
		log.Fatalf("找不到前端目录 web（可用 -web 指定路径）")
	}

	b := newBackend(src, workDir)

	mux := http.NewServeMux()
	mux.Handle("/api/register", handleOperation(b, 1))
	mux.Handle("/api/login", handleOperation(b, 2))
	mux.Handle("/api/exit", handleOperation(b, 3))
	mux.Handle("/api/status", handleStatus(b))
	mux.Handle("/", http.FileServer(http.Dir(web)))

	url := "http://" + *addr + "/"
	log.Printf("后端源码（只读运行）：%s", src)
	log.Printf("前端目录：%s", web)
	log.Printf("桥接服务已启动：%s", url)
	log.Printf("接口：POST /api/register  POST /api/login  POST /api/exit  GET /api/status")

	if *open && strings.HasPrefix(*addr, "127.0.0.1") {
		go func() {
			time.Sleep(500 * time.Millisecond)
			_ = exec.Command("cmd", "/c", "start", "", url).Start()
		}()
	}

	if err := http.ListenAndServe(*addr, withCORS(mux)); err != nil {
		log.Fatalf("监听 %s 失败：%v", *addr, err)
	}
}
