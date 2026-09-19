
#include <iostream>
#include<vector>
#include<algorithm>
#include<bitset>
#include<cstdio>
using namespace std;
//第一题
void test1() {
    cout << "*********************************\n";
    cout << "Lab Task2\n";
    cout << "\n";
    cout << "Procedural Programming\n";
    cout << "......\n" << endl;;
    //感谢马里奥打印提醒
}
//第二题
void test2() {
    cout << "请输入取款金额>";
    int sum,hun,ten,digit;
    cin >> sum;
    hun = sum / 100;
    ten = (sum - hun*100) / 10;
    digit = (sum - hun*100 - ten*10);
    cout << sum << "=" << hun << "*100+" << ten << "*10+" << digit << endl;
}
//第三题
void test3() {
    cout << "输入数字展示斐波那契数列前n项>";
    int a,sum,f1=0,f2=1;
    cin >> a;
    vector<int>arr;
    arr.push_back (f1);
    arr.push_back(f2);
    for (int i=0;i<a+1;i++){
        sum = arr[i] + arr[i + 1];
        arr.push_back(sum);
    }
    arr.pop_back();
    arr.pop_back();
    arr.pop_back();
    //因该是角标错了，每次都多三个数出来，选择这种笨办法解决
    for (int j = 0; j < arr.size(); j++) { cout << arr[j] << endl; }
}
//第四题第一步
void test41() {
    int a, b;
    cin >> a >> b;
    vector <int>arr{1};
    for (int i = 1; i <= a && i <= b; i++) {
        if (a % i == 0 && b % i == 0&&i!=1) {
            arr.push_back(i);
        }
    }
    sort(arr.begin(), arr.end(), greater<int>());
    cout << arr[0];
}
void test42() {
    int a, b;
    cin >> a >> b;
    vector<int>arr{a*b};
    for(int i=a*b;i>=1;i--){
        if (i % a == 0 && i % b == 0) {
            arr.push_back(i);
        }
    }
    sort(arr.begin(), arr.end());
    cout << arr[0];
}
//第五题
void test5() {
    vector<int>arr;
    for (int i = 0; i < 10; i++) {
        int a;
        cin >> a;
        arr.push_back(a);
    }
    double sum=0,avg;
    for (int j = 0; j < arr.size(); j++) {
        sum = arr[j]+sum;
    }
    avg = sum / 10;
    cout << "平均数为>" << avg << endl;
    sort(arr.begin(), arr.end());
    cout << "max=" << arr[9]<<"  " << "min=" << arr[0];
}
//第六题（没思路）
/*void test6() {
    vector<string>arr;
    string fst{ "*" };
    for (int i = 0; i <= 7; i++) {
       
    }
}*/
//第七题
void test7() {
    int a;
    cin >> a;
    cout << bitset<32>(a) << endl;;
    vector<int>arr;
    while (a>0) {
        int bit = a % 2;
        arr.push_back(bit);
        a = a / 2;
    }
    reverse(arr.begin(), arr.end());
    for (int i = 0; i< arr.size(); i++)
        cout << arr[i];//只会vector是这样的
}
int main()
{
    test1();
    test2();
    test3();
    test41();
    test42();
    test5();
    //test6();
    test7();
}


