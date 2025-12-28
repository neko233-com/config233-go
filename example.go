package main

import (
	"fmt"
	"reflect"
	"time"

	"config233-go/pkg/config233"
	"config233-go/pkg/config233/json"
	"config233-go/pkg/config233/tsv"
)

func main() {
	fmt.Println("Config233-Go 示例")

	// 创建配置实例
	cfg := config233.NewConfig233().
		Directory("./config").
		AddConfigHandler("json", &json.JsonConfigHandler{}).
		AddConfigHandler("tsv", &tsv.TsvConfigHandler{}).
		RegisterConfigClass("Student", reflect.TypeOf(Student{})).
		Start()

	// 获取配置
	students := cfg.GetConfigList(reflect.TypeOf(Student{}))
	fmt.Printf("加载的学生数量: %d\n", len(students.([]interface{})))

	// 注册热更新对象
	updater := &StudentUpdater{}
	cfg.RegisterForHotUpdate(updater)

	fmt.Println("配置加载完成，监听文件变化...")
	fmt.Println("你可以修改 config/Student.json 文件来测试热更新")
	fmt.Println("按 Ctrl+C 退出")

	// 保持运行
	for {
		time.Sleep(time.Second)
	}
}

// Student 配置结构体
type Student struct {
	ID   int    `json:"id" config233:"uid"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

// StudentUpdater 热更新处理器
type StudentUpdater struct {
	StudentMap map[int]*Student `config233:"inject"`
}

func (u *StudentUpdater) UpdateStudents() {
	if u.StudentMap != nil {
		fmt.Printf("学生映射已更新，当前学生数: %d\n", len(u.StudentMap))
		for id, student := range u.StudentMap {
			fmt.Printf("ID: %d, Name: %s, Age: %d\n", id, student.Name, student.Age)
		}
	}
}
