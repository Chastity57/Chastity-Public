#include <iostream>
#include<vector>
#include<string>
#include<algorithm>
#include<ranges>
using namespace std;
//第一题
void test1() {
    while (true) {
        vector<int>arr{ 3,2,1,5,6,4,5,4,9 };
        sizeof(arr);
        cout << sizeof(arr) << endl;
        cout << arr.size();
        sort(arr.begin(), arr.end());
        auto rep = unique(arr.begin(), arr.end());
        arr.erase(rep, arr.end());
        cout << "输入查看数组中第k大的元素" << endl;
        int k;
        cin >> k;
        cout << arr[k - 1] << endl;
    }
}
//第二题
void test2(){
    vector<int>nums1{ 3,5,6,8,1,2,0,0,0,0 };
    vector<int>nums2{ 1,2,6 };
    int m = 6;
    int n = 3;
    sort(nums1.begin(), nums1.end());
    for (int i = 0; i < n; i++) {
        nums1[i] = nums2[i];
    }
    sort(nums1.begin(), nums1.end());
    for (int j = 0; j < 10; j++) {
        cout << nums1[j]<<" ";
    }
}
//第三题
void test3() {
    vector<int>num1;
    for (int i = 1; i < 20251018; i++)
    {
        int temp = i;
        int sum = 0;
        while (temp > 0) {
            sum += temp % 10;
            temp /= 10;
        }
        if (sum % 5 == 0) { num1.push_back(i); };
    }
    cout << num1.size();
}
//第四题
void test4() {
 vector<int>num;
 for (int i = 1; i <10000; i++) {
     int temp1 = i;
     vector<int>numtemp;
     while (temp1 > 0) {
         numtemp.push_back(temp1 % 10);
         temp1 /= 10;
     }
     vector<int>num0, num2, num5;
     for (int j = 0; j < numtemp.size(); j++) {
         if (numtemp[j] == 0) { num0.push_back(j); }
         if (numtemp[j] == 2) { num2.push_back(j); }
         if (numtemp[j] == 5) { num5.push_back(j); }
     }
         if (num0.size() > 0 && num2.size() > 1 && num5.size() > 0) {
             num.push_back(i);
         }
 }
 cout << num.size();
}
//第五题
int digitsum(int num) {
    int sum = 0;
    while (num> 0) {
        sum = num % 10;
        num /= 10;
    }
    return sum;
}
void test5() {
    vector<int>suc;
    for (int year = 1900; year < 9999; year++) {
        int sumy = digitsum(year);
        for (int month = 1; month < 13; month++) {
            int summ = digitsum(month);
            for (int date = 1; date < 31; date++) {
                int sumd = digitsum(date);
                if (sumy == summ + sumd) { suc.push_back(year); }
            }
        }
    }
    cout << suc.size();
}
//第六题
vector<int>arr;
void prinum() {
    for (int num = 2; num < 2021; num++) {
        bool ok = true;
        for (int div = 2; div *div<= num; div++) {
            if (num % div == 0) {
                ok = false;
                break;
            }
        }
        if (ok) {
            arr.push_back(num);
        }
    }
}
bool truepri(int num) {
    return(num == 2 || num == 3 || num == 5 || num == 7);
}
void test6() {
    vector<int>trueprinum;
    prinum();
    for (int i = 0; i < arr.size(); i++) {
        int num;
        num = arr[i];
        bool ok = true;
        while (num > 0) {
            int digit=num % 10;
            if (!truepri(digit)) {
                ok = false;
                break;
            }
            num /= 10;
        }
        if (ok) {
            trueprinum.push_back(arr[i]);
        }
    }
    cout << trueprinum.size();
}
int main()
{
    test1();
    test2();
    test3();
    test4();
    test5();
    test6();
}

