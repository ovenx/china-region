package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type Province struct {
	Name   string `json:"name"`
	Code   string `json:"code"`
	Cities []City `json:"cities"`
}

type City struct {
	Name  string `json:"name"`
	Code  string `json:"code"`
	Areas []Area `json:"areas"`
}

type Area struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type DataLine struct {
	Name  string
	Code  string
	Level int
}

const Link = "http://www.mca.gov.cn/article/sj/xzqh/2020/2020/2020112010001.html"

func main() {

	// get data
	fmt.Println("start get data")
	dataLine := GetDataLine()

	// format data
	data := GetFormatedData(dataLine)

	// create json file
	fmt.Println("start create json file")
	jsonData, _ := json.MarshalIndent(&data, "", "  ")
	WriteToFile("region.json", string(jsonData))

	// create sql file
	fmt.Println("start create sql file")
	sql := "DROP TABLE IF EXISTS `region`;\n" +
		"CREATE TABLE `region` (" +
		"`id` int(11) NOT NULL AUTO_INCREMENT," +
		"`code` int(11) NOT NULL COMMENT '行政代码'," +
		"`name` varchar(50) NOT NULL COMMENT '名称'," +
		"`level` tinyint(4) NOT NULL COMMENT '级别：1.省 2.市 3.区(县)'," +
		"`parent` int(11) NOT NULL COMMENT '上级行政代码'," +
		"PRIMARY KEY (`id`)," +
		"UNIQUE KEY `code` (`code`)" +
		") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='地区表';\n\n"

	for _, proItem := range data {
		sql += fmt.Sprintf("INSERT INTO `region`(`code`, `name`, `level`, `parent`) VALUES('%s', '%s', '%s', '%s');\n", proItem.Code, proItem.Name, "1", "0")
		for _, cityItem := range proItem.Cities {
			sql += fmt.Sprintf("INSERT INTO `region`(`code`, `name`, `level`, `parent`) VALUES('%s', '%s', '%s', '%s');\n", cityItem.Code, cityItem.Name, "2", proItem.Code)
			for _, areaItem := range cityItem.Areas {
				sql += fmt.Sprintf("INSERT INTO `region`(`code`, `name`, `level`, `parent`) VALUES('%s', '%s', '%s', '%s');\n", areaItem.Code, areaItem.Name, "3", cityItem.Code)
			}
		}
	}
	WriteToFile("region.sql", sql)
	fmt.Println("done!")
}

func GetFormatedData(dataLine []DataLine) []Province {
	data := make([]Province, 0)
	province := Province{}
	city := City{}
	area := Area{}
	for _, item := range dataLine {
		if item.Level == 1 {
			if (!reflect.DeepEqual(province, Province{})) {
				if !reflect.DeepEqual(city, City{}) {
					province.Cities = append(province.Cities, city)
				}
				data = append(data, province)
			}
			province = Province{Code: item.Code, Name: item.Name, Cities: []City{}}
			city = City{}
			area = Area{}
		} else if item.Level == 2 {
			if !reflect.DeepEqual(city, City{}) {
				province.Cities = append(province.Cities, city)
			}
			city = City{Code: item.Code, Name: item.Name, Areas: []Area{}}
			area = Area{}
		} else {
			area = Area{Code: item.Code, Name: item.Name}
			city.Areas = append(city.Areas, area)
		}
	}
	data = append(data, province)
	return data
}

func GetDataLine() []DataLine {
	res, err := http.Get(Link)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	dataLine := make([]DataLine, 0)
	line := DataLine{}
	doc.Find("tr[height=\"19\"]").Each(func(i int, s *goquery.Selection) {
		code := ""
		name := ""
		s.Find("td").Each(func(j int, s *goquery.Selection) {
			if j == 1 {
				code = s.Text()
			} else if j == 2 {
				strEncode := url.QueryEscape(s.Text())
				strEncode = strings.Replace(strEncode, "%C2%A0", "%20", -1)
				// strEncode = strings.Replace(strEncode, "+", "", -1)
				name, _ = url.QueryUnescape(strEncode)
			} else {
				return
			}
		})
		filterName := strings.Replace(name, " ", "", -1)
		if name[:1] != " " {
			line = DataLine{Code: code, Name: filterName, Level: 1}
			dataLine = append(dataLine, line)
		} else if name[1:2] != " " {
			line = DataLine{Code: code, Name: filterName, Level: 2}
			dataLine = append(dataLine, line)
		} else {
			if line.Level == 1 {
				pCode, _ := strconv.Atoi(line.Code)
				line = DataLine{Code: strconv.Itoa(pCode + 100), Name: line.Name, Level: 2}
				dataLine = append(dataLine, line)
			}
			line = DataLine{Code: code, Name: filterName, Level: 3}
			dataLine = append(dataLine, line)
		}
	})
	return dataLine
}

func WriteToFile(filePath string, str string) {
	out, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println("open file fail:", err)
		return
	}
	defer out.Close()
	writer := bufio.NewWriter(out)
	writer.WriteString(str)
	writer.Flush()
}
