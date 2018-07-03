package main

import (
	"github.com/gomodule/redigo/redis"
	"fmt"
)

func TestString(c redis.Conn) {
	// string 字符串
	_, err := c.Do("SET", "fuck", "shit")
	if err != nil {
		panic(err)
	}

	shit, err := redis.String(c.Do("GET", "fuck"))
	if err != nil {
		panic(err)
	}

	fmt.Println(shit)
}

func TestHash(c redis.Conn) {
	// hash 哈希

	// 方式1: 结构体 reactor

	var p1, p2 struct {
		Title  string `redis:"title"`
		Author string `redis:"author"`
		Body   string `redis:"body"`
	}

	p1.Title = "fuck title"
	p1.Author = "shit author"
	p1.Body = "sick body"

	if _, err := c.Do("HMSET", redis.Args{}.Add("id1").AddFlat(&p1)...); err != nil {
		panic(err)
	}

	// 方式2: map (我认为最多就是1. map嵌套字符串 2. map嵌套map)
	m := map[string]string{
		"title":  "fuck title2",
		"author": "shit author2",
		"body":   "sick body2",
	}

	if _, err := c.Do("HMSET", redis.Args{}.Add("id2").AddFlat(m)...); err != nil {
		panic(err)
	}

	for _, id := range [] string{"id1", "id2"} {

		v, err := redis.Values(c.Do("HGETALL", id))
		if err != nil {
			panic(err)
		}

		if redis.ScanStruct(v, &p2); err != nil {
			panic(err)
		}

		fmt.Printf("%+v\n", p2)
	}
}

func TestList(c redis.Conn) {
	c.Send("DEL", "albums")

	c.Send("HMSET", "album:1", "title", "Red", "rating", 5)
	c.Send("HMSET", "album:2", "title", "Earthbound", "rating", 1)
	c.Send("HMSET", "album:3", "title", "Beat")

	c.Send("LPUSH", "albums", "1")
	c.Send("LPUSH", "albums", "2")
	c.Send("LPUSH", "albums", "3")

	values, err := redis.Values(c.Do("SORT", "albums",
		"BY", "album:*->rating",
		"GET", "album:*->title",
		"GET", "album:*->rating"))

	if err != nil {
		panic(err)
	}

	for len(values) > 0 {
		var title string
		rating := -1
		values, err = redis.Scan(values, &title, &rating)
		if err != nil {
			panic(err)
		}

		if rating == -1 {
			fmt.Println(title, "not-rated")
		} else {
			fmt.Println(title, rating)
		}

	}
}

func TestSet(c redis.Conn) {
	c.Do("SADD", "set_with_integers", 4, 5, 6)

	ints, _ := redis.Ints(c.Do("SMEMBERS", "set_with_integers"))
	fmt.Printf("%#v\n", ints)
}

func TestZSet(c redis.Conn) {
	// TODO

}

func main() {
	c, err := redis.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		panic(err)
	}
	defer c.Close()

	TestString(c)

	TestHash(c)

	TestList(c)

	TestSet(c)

	TestZSet(c)
}
