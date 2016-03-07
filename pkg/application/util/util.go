package util

import (
	"errors"
	applicationapi "github.com/openshift/origin/pkg/application/api"
	"strings"
)

func Contains(arr []string, str string) bool {
	for _, ele := range arr {
		if ele == str {
			return true
		}
	}

	return false
}

func Parse(items string) (applicationapi.ItemList, error) {
	list := applicationapi.ItemList{}
	arr := strings.Split(items, ",")

	for _, s := range arr {
		item := strings.Split(strings.TrimSpace(s), "=")
		if len(item) != 2 {
			return nil, errors.New("items wrong format")
		}

		list = append(list, applicationapi.Item{
			Kind: item[0],
			Name: item[1],
		})
	}

	return list, nil
}
