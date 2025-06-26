package service

import (
	"goserver/internal/web/stats/app/entity"

	"gopkg.in/mgo.v2/bson"
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

func (this *systemService) GetAgentList(isParent bool) map[string][]entity.Perm {
	var list []entity.Perm
	m := bson.M{}
	if !isParent {
		m["parent"] = ""
	}
	// m["module"] = bson.M{"$in": []string{"user", "statistics"}}
	// m["action"] = bson.M{"$in": []string{"datalist", "share"}}

	m = bson.M{
		"$or": []bson.M{
			bson.M{
				"$and": []bson.M{
					bson.M{"module": "statistics"},
					bson.M{"action": bson.M{"$in": []string{"datalist", "share"}}},
				},
			},
			bson.M{"module": bson.M{"$in": []string{"role", "user"}}, "parent": ""},
		},
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
