--- 功能描述：前提windows系统安装cygwin，并在5000端口上启动sockpro服务，接收GET请求
--- get请求参数为dirid
--- 1、先解压该目录下的zip文件
--- 2、启动计算进程VTK2GLTF.exe，将计算结果文件输出到该目录的result文件夹下
--- 3、返回计算进程的结果
--- Created by yangpengfei.
--- DateTime: 2019/8/14 9:30
---
local shell = require("resty.shell")
ngx.header.content_type="application/json;charset=utf-8"
-- 读取get请求参数
local local_args = ngx.req.get_uri_args()
-- dirid参数传入所在目录名称
local dirid = local_args["dirid"]
if nil == dirid then
    ngx.print("dirid参数不能为空！")
else
    local args = {
        -- socket = "unix:/tmp/shell.sock",
        -- 先在此端口上启动sockpro服务 ./sockproc.exe 5000 --foreground
        socket = {host = "127.0.0.1", port = 5000},
        data = "\r\n",
    }
    ---- 解压文件
    --local status0, result0, err0 = shell.execute("7z x `cygpath -w /cygdrive/d/*.zip` -y -o`cygpath -w /cygdrive/d`", args)
    --ngx.log(ngx.INFO, "unzip status: ".. status0)
    --ngx.log(ngx.INFO, "unzip result: ".. result0)
    --if status0 ~= 0 then
    --    ngx.say("err: unzip 解压文件异常")
    --end

    -- 启动计算进程
    local exeDir = "/cygdrive/d/openform-web/VTKRelease/VTK2GLTF.exe"
    local vtkDir = " `cygpath -w /cygdrive/d/openform-web/VTKRelease/"..dirid.."` "
    local GLTFDir = " `cygpath -w  /cygdrive/d/openform-web/VTKRelease/"..dirid.."/result".."` "
    local GDAL_DATADir = " `cygpath -w /cygdrive/d/openform-web/VTKRelease/GDAL_DATA` "
    local cmd = exeDir .. vtkDir .. GLTFDir .. GDAL_DATADir

    --local status, result, err = shell.execute("/cygdrive/d/openform-web/VTKRelease/VTK2GLTF.exe `cygpath -w /cygdrive/d/openform-web/VTKRelease/data` `cygpath -w  /cygdrive/d/openform-web/VTKRelease/data` `cygpath -w /cygdrive/d/openform-web/VTKRelease/GDAL_DATA`", args)
    local status, result, err = shell.execute(cmd, args)

    if err ~= nil then
        ngx.say("err: ".. err)
    end
    if result ~= nil then
        ngx.say("result: ".. result)
    end
end