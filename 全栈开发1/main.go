// 要求1，对账号唯一性进行判断
// 2，密码用哈希加密存储
// 3，登录时对比哈希值
// 4，无论登录成功与否返回对应信息
package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
)

// 存储用户信息
type Users /*创建名为user的结构体*/ struct {
	username string
	password string
	salt     string
}

// 全局变量创建已注册用户存储
var userlist []Users //结构体不能遍历，创建变量方便遍历
var ok bool

// 密码加密
func hashpass(salt, password string) string {
	sum := sha256.Sum256([]byte(salt + password))
	return hex.EncodeToString(sum[:])
}
func genSalt() string {
	b := make([]byte, 16) // 16字节 = 128位随机盐
	_, err := rand.Read(b)
	if err != nil {
		// 随机源失败时使用时间作为降级方案（正常情况下不会发生）
		return hex.EncodeToString([]byte(fmt.Sprintf("%d", os.Getpid())))
	}
	return hex.EncodeToString(b)
}

// 注册操作
func register(username, password string) {
	var s Users
	s.username = username
	s.password = hashpass(password, s.salt)
	fmt.Println("注册成功")
	userlist = append(userlist, s)
}

// 登录操作
func login(username, password string) {
	for _, user := range userlist {
		if user. /*调用结构体*/ username /*取出结构体中的用户名进行对比*/ == username {
			if user.password == hashpass(password, user.salt) {
				fmt.Println("登录成功")
				return
			} else {
				fmt.Println("密码错误")
				return
			}
		}
	}
	fmt.Println("用户不存在")
}

// 用户查重及用户不存在应对
func checkname(username string) bool {
	for _, user := range userlist {
		if user.username == username {
			fmt.Println("该用户名已被注册")
			loginact(0) //退回list列表
			ok = true
			return ok
		}
	}
	ok = false
	return ok
}

// 用户操作界面
func loginact(userorder int) bool {
	if userorder == 0 {
		ok = false
		return ok
	}
	if userorder == 1 {
		fmt.Println("注册")
		var username, password string
		fmt.Print("输入你的用户名>")
		fmt.Scan(&username)
		if checkname(username) {
			ok = false
			return ok
		}
		fmt.Print("输入密码>")
		fmt.Scan(&password)
		register(username, password)
		ok = false
		return ok
	} else if userorder == 2 {
		fmt.Println("登录")
		var username, password string
		fmt.Print("输入你的用户名>")
		fmt.Scan(&username)
		fmt.Print("输入密码>")
		fmt.Scan(&password)
		login(username, password)
	} else if userorder == 3 {
		fmt.Println("退出登录")
		os.Exit(0)
	}
	ok = false
	return ok
}

// loginact返回值处理
func actreturn(ok bool) {
	switch ok {
	case true:
		return
	case false:
		break
	}
}
func main() {
	for {
		fmt.Println("==========home list==========")
		fmt.Println("你想做什么")
		fmt.Printf("1,注册账号\t2,登录账号、\t3,退出程序\n")
		var userorder int
		fmt.Scan(&userorder)
		switch userorder {
		case 1:
			loginact(1)
			actreturn(ok)
		case 2:
			loginact(2)
			actreturn(ok)
		case 3:
			loginact(3)
			actreturn(ok)
		default:
			fmt.Println("错误命令")
			var sweep string
			fmt.Scanln(&sweep)
			continue
		}
	}
}
