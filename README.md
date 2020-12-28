## 中国省市区数据

自动爬取最新的省市区数据，生成 json 和 sql 文件

最新2020年10月份数据，数据来源于[中华人民共和国民政部](http://www.mca.gov.cn/article/sj/xzqh/2020/2020/2020112010001.html)

如何使用：
```bash
# 修改 link 为最新的链接
const Link = "http://www.mca.gov.cn/article/sj/xzqh/2020/2020/2020112010001.html"
go run main.go
```