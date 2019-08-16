--- 功能描述：前提启动file-manager服务，接收body类型为form-data的PUT请求
--- 1、先在当前目录创建一个以GUID作为名称的文件夹
--- 2、调用内部代理接口（/uploadAndUnzip）把请求中的文件(一个或多个)上传到此文件夹，并解压
--- 3、若上传成功，返回文件夹GUID；若上传失败，删除第1步所创建的文件夹
--- Created by yangpengfei.
--- DateTime: 2019/8/15 9:30
---
local guid = require ("guid")
local http = require("resty.http")

local fileManager = "http://127.0.0.1:8081/"
local httpc = http:new()

-- 创建新文件夹
-- 生成一个随机标识，作为新文件夹名称
local dirName = guid.generate()
local res1, err1 = httpc:request_uri(
        fileManager,
        {
            path = fileManager .. "?format=json",
            method = "POST",
            headers = {
                ["Content-Type"] = "application/json;charset=UTF-8",
            },
            body = "{\"action\":\"createFolder\",\"params\":{\"source\":\"/"..dirName.."\"}}"
        }
)
--若文件夹创建失败，则返回状态码并退出
if res1.status ~= ngx.HTTP_OK then
    ngx.exit(res1.status)
end

--若文件夹创建成功，则上传文件，并解压文件
ngx.req.read_body()
local res = ngx.location.capture('/uploadAndUnzip',
        { method = ngx.HTTP_PUT,
          args = {unzip = "true", destPath = dirName},
          always_forward_body = true }
)

--如果上传失败，删除所建目录
if nil == res then
    local res2, err2 = httpc:request_uri(
            fileManager,
            {
                path = fileManager,
                method = "POST",
                headers = {
                    ["Content-Type"] = "application/json;charset=UTF-8",
                },
                --{"action":"delete","paramslist":["/7AED6B57-18E1-B312-77E9-B756AE4D65F9"]}
                body = "{\"action\":\"delete\",\"paramslist\":[\"/"..dirName.."\"]}"
            }
    )
    ngx.exit(res2.status)
end

if res.status ~= ngx.HTTP_OK then
    ngx.exit(res.status)
end

--创建result文件夹
local res3, err3 = httpc:request_uri(
        fileManager,
        {
            path = fileManager .. "?format=json",
            method = "POST",
            headers = {
                ["Content-Type"] = "application/json;charset=UTF-8",
            },
            body = "{\"action\":\"createFolder\",\"params\":{\"source\":\"/"..dirName.."/result\"}}"
        }
)
-- res.body="ok"
ngx.print("上传文件所在目录为："..dirName)