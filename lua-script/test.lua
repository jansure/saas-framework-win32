local utf8 = require('lua-utf8')
local s1 = "aaa中文"
print("s1的长度："..utf8.len(s1))
print(utf8.insert(s1, 3, "li里"))
print(utf8.insert(s1, 0, "li里"))
print(utf8.insert(s1, -1, "li里"))
print(utf8.insert(s1, -utf8.len(s1), "li里"))
print(utf8.fold(s1))
print(utf8.upper("aaa中"))
print(utf8.next("asdag"))
for pos, code in utf8.next, "asdag中文" do
   print("处于第"..pos.."位的字符Unicode码为: "..code)
end

-- Unicode码转义为utf-8编码
local u = utf8.escape
print(u"%123%u123%{123}%u{123}%xABC%x{ABC}")
print(u"%20013%25991%%123%?%d%%u%27012")
print(utf8.escape("%123%u123%{123}%u{123}%xABD%x{ABD}"))


print(utf8.reverse("abc中d"))
