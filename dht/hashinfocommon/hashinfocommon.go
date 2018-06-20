package hashinfocommon

import (
	"errors"
	"reflect"
	"net"
	"strconv"
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

func CheckDictInfoHashValid(dict map[string]interface{}) error {
	if infohash, ok := dict["info_hash"]; !ok ||
		reflect.TypeOf(infohash).Kind() != reflect.String ||
		len(infohash.(string)) != 20 {
		return errors.New("error id")
	}
	return nil
}

func GetObjWithCheck(b interface{}) (map[string]interface{}, error) {
	if obj, ok := b.(map[string]interface{}); !ok {
		return nil, errors.New("not an obj")
	} else {

		// transaction id
		if t, ok := obj["t"]; !ok || reflect.TypeOf(t).Kind() != reflect.String {
			return nil, errors.New("error key \"t\"")
		}

		// msg type
		if y, ok := obj["y"]; !ok || reflect.TypeOf(y).Kind() != reflect.String {
			return nil, errors.New("error key \"y\"")
		} else {

			// req
			if y == "q" {

				// req type
				if q, ok := obj["q"]; !ok || reflect.TypeOf(q).Kind() != reflect.String {
					return nil, errors.New("error key \"q\"")
				}

				// req parameter
				if a, ok := obj["a"]; !ok || reflect.TypeOf(a).Kind() != reflect.Map {
					return nil, errors.New("error key \"a\"")
				}
			}

			// response
			if y == "r" {
				// response parameter
				if r, ok := obj["r"]; !ok || reflect.TypeOf(r).Kind() != reflect.Map {
					return nil, errors.New("error key \"r\"")
				} else {
					if id, ok := r.(map[string]interface{})["id"]; !ok || reflect.TypeOf(id).Kind() != reflect.String || len(id.(string)) != 20 {
						return nil, errors.New("error id \"r\"")
					}
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

	if nodes, ok := r["nodes"]; !ok ||
		reflect.TypeOf(nodes).Kind() != reflect.String ||
		len(nodes.(string))%26 != 0 {
		return errors.New("error r.nodes")
	}

	return nil
}

func CheckReqPingValid(req map[string]interface{}) error {
	a := req["a"].(map[string]interface{})
	return CheckDictIdValid(a)
}

func CheckReqGetPeersValid(req map[string]interface{}) error {
	a := req["a"].(map[string]interface{})
	return CheckDictInfoHashValid(a)
}

func CheckReqAnnouncePeerValid(req map[string]interface{}) error {
	a := req["a"].(map[string]interface{})

	if err := CheckDictInfoHashValid(a); err != nil {
		return err
	}

	if port, ok := a["port"]; !ok ||
		reflect.TypeOf(port).Kind() != reflect.Int ||
		port.(int) <= 0 || port.(int) > 65535 {
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
		port := int(uint16(str[p])<<8 | uint16(str[p+1]))

		if port <= 0 || port > 65535 {
			continue
		}

		node.Addr = ip + ":" + strconv.Itoa(port)

		nodes = append(nodes, node)
	}

	return nodes
}
