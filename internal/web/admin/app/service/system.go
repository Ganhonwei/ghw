package service

import (
	"goserver/internal/web/admin/app/entity"

	"github.com/globalsign/mgo/bson"
)

type systemService struct{}

func (this *systemService) GetPermList(isParent bool) map[string][]entity.Perm {
	var list []entity.Perm
	m := bson.M{}
	if !isParent {
		m["parent"] = ""
	}
	ListByQ(Perms, m, &list)

	result := make(map[string][]entity.Perm)
	for _, v := range list {
		v.Key = v.Module + "." + v.Action
		if _, ok := result[v.Module]; !ok {
			result[v.Module] = make([]entity.Perm, 0)
		}
		result[v.Module] = append(result[v.Module], v)
	}
	return result
}

func (this *systemService) GetPermChild(action string) []entity.Perm {
	var list []entity.Perm
	m := bson.M{}
	m["parent"] = action
	ListByQ(Perms, m, &list)

	// result := make(map[string][]entity.Perm)
	// for _, v := range list {
	// 	v.Key = v.Module + "." + v.Action
	// 	if _, ok := result[v.Module]; !ok {
	// 		result[v.Module] = make([]entity.Perm, 0)
	// 	}
	// 	result[v.Module] = append(result[v.Module], v)
	// }
	return list
}
