###
 # @Author: Strong
 # @Date: 2025-03-02 21:23:00
 # @LastEditors: Strong
 # @LastEditTime: 2025-03-02 21:24:15
 # @Description: 请填写简介
### 


GOOS=windows GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o ./build/windows.exe