 ###
 # @Author: Chris
 # @Date: 2025-03-02 21:15:07
 # @LastEditors: Chris
 # @LastEditTime: 2025-03-02 21:21:42
 # @Description: 请填写简介
### 


GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o ./build/linux