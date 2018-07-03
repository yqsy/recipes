package main

import (
	"github.com/go-redis/redis"
	"fmt"
)

func main() {
	client := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,})

	pong, err := client.Ping().Result()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%v\n", pong)

	if err = client.Set("fuck", "shit", 0).Err(); err != nil {
		panic(err)
	}

	val, err := client.Get("fuck").Result()
	if err != nil {
		panic(err)
	}

	fmt.Printf("fuck's value equals %v\n", val)

	val2, err := client.Get("fuckuuuuu").Result()
	if err == redis.Nil {
		fmt.Printf("relax the key's value is not exist\n")
	} else if err != nil {
		panic(err)
	} else {
		fmt.Printf("fuckuuuuu: %v\n", val2)
	}

}
