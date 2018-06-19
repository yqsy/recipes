package hashinfocommon

import (
	"errors"
	"reflect"
	"net"
	"strconv"
)

const (
	TokenLen = 2
)

type Node struct {
	Id   string
	Addr string
}

func CheckDictIdValid(dict map[string]interface{}) error {
	if id, ok := dict["id"]; !ok ||
		reflect.TypeOf(id).Kind() != reflect.String ||
		len(id.(string)) != 20 {
		return errors.New("error id")
	}
	return nil
}

func GetObjWithCheck(b interface{}) (map[string]interface{}, error) {
	if obj, ok := b.(map[string]interface{}); !ok {
		return nil, errors.New("not an obj")
	} else {
		if t, ok := obj["t"]; !ok || reflect.TypeOf(t).Kind() != reflect.String {
			return nil, errors.New("error key \"t\"")
		}

		if y, ok := obj["y"]; !ok || reflect.TypeOf(y).Kind() != reflect.String {
			return nil, errors.New("error key \"y\"")
		} else {

			// req
			if y == "q" {
				if q, ok := obj["q"]; !ok || reflect.TypeOf(q).Kind() != reflect.String {
					return nil, errors.New("error key \"q\"")
				}
				if a, ok := obj["a"]; !ok || reflect.TypeOf(a).Kind() != reflect.Map {
					return nil, errors.New("error key \"a\"")
				}
			}

			// response
			if y == "r" {
				if r, ok := obj["r"]; !ok || reflect.TypeOf(r).Kind() != reflect.Map {
					return nil, errors.New("error key \"r\"")
				}
			}

			// error
			if y == "e" {
				if e, ok := obj["e"]; !ok || reflect.TypeOf(e).Kind() != reflect.Slice {
					return nil, errors.New("error key \"e\"")
				} else {

					E := e.([]interface{})

					if len(E) != 2 ||
						reflect.TypeOf(E[0]).Kind() != reflect.Int ||
						reflect.TypeOf(E[1]).Kind() != reflect.String {
						return nil, errors.New("obj err msg error")
					}
				}
			}
		}
		return obj, nil
	}
}

func CheckResFindNodeValid(res map[string]interface{}) error {
	r := res["r"].(map[string]interface{})

	if err := CheckDictIdValid(r); err != nil {
		return err
	}

	if nodes, ok := r["nodes"]; !ok ||
		reflect.TypeOf(nodes).Kind() != reflect.String ||
		len(nodes.(string))%26 != 0 {
		return errors.New("error r.nodes")
	}

	return nil
}

func CheckReqPingValid(req map[string]interface{}) error {
	a := req["a"].(map[string]interface{})

	if err := CheckDictIdValid(a); err != nil {
		return err
	}

	return nil
}

func CheckReqGetPeersValid(req map[string]interface{}) error {
	a := req["a"].(map[string]interface{})

	if hashinfo, ok := a["info_hash"]; !ok ||
		reflect.TypeOf(hashinfo).Kind() != reflect.String ||
		len(hashinfo.(string)) < TokenLen {
		return errors.New("error a.hash_info")
	}

	return nil
}

func CheckReqAnnouncePeerValid(req map[string]interface{}) error {
	a := req["a"].(map[string]interface{})

	if hashinfo, ok := a["info_hash"]; !ok ||
		reflect.TypeOf(hashinfo).Kind() != reflect.String ||
		len(hashinfo.(string)) != 20 {
		return errors.New("error a.hash_info")
	}

	if port, ok := a["port"]; !ok ||
		reflect.TypeOf(port).Kind() != reflect.Int ||
		port.(int) < 0 || port.(int) > 65535 {
		return errors.New("error a.port")
	}

	return nil
}

func GetNodes(str string) []Node {
	var nodes []Node

	for i := 0; i < len(str)/26; i++ {
		p := i * 26

		if p+25 >= len(str) {
			break
		}

		var node Node
		node.Id = str[p : p+20]
		p += 20
		ip := net.IPv4(str[p],
			str[p+1],
			str[p+2],
			str[p+3]).String()
		p += 4
		port := strconv.Itoa(int(uint16(str[p])<<8 | uint16(str[p+1])))

		node.Addr = ip + ":" + port

		nodes = append(nodes, node)
	}

	return nodes
}
