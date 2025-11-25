package singleton

import (
	"cmp"
	"slices"

	"github.com/nezhahq/nezha/model"
	"github.com/nezhahq/nezha/pkg/utils"
)

type AutoSSHClass struct {
	class[uint64, *model.AutoSSH]
}

func NewAutoSSHClass() *AutoSSHClass {
	var sortedList []*model.AutoSSH

	DB.Find(&sortedList)
	list := make(map[uint64]*model.AutoSSH, len(sortedList))
	for _, mapping := range sortedList {
		list[mapping.ID] = mapping
	}

	return &AutoSSHClass{
		class: class[uint64, *model.AutoSSH]{
			list:       list,
			sortedList: sortedList,
		},
	}
}

func (c *AutoSSHClass) Update(a *model.AutoSSH) {
	c.listMu.Lock()
	c.list[a.ID] = a
	c.listMu.Unlock()
	c.sortList()
}

func (c *AutoSSHClass) Delete(idList []uint64) {
	c.listMu.Lock()
	for _, id := range idList {
		delete(c.list, id)
	}
	c.listMu.Unlock()
	c.sortList()
}

func (c *AutoSSHClass) sortList() {
	c.listMu.RLock()
	defer c.listMu.RUnlock()

	sortedList := utils.MapValuesToSlice(c.list)
	slices.SortFunc(sortedList, func(a, b *model.AutoSSH) int {
		return cmp.Compare(a.ID, b.ID)
	})

	c.sortedListMu.Lock()
	defer c.sortedListMu.Unlock()
	c.sortedList = sortedList
}
