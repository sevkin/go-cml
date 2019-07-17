package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/sevkin/go-cml/xml"
	"github.com/vrischmann/envconfig"
)

type (
	_offer struct {
		*xml.Предложение
	}

	_product struct {
		*xml.Товар
		// offers map[string]
	}

	_group struct {
		*xml.Группа
		path     []string
		products map[string]*_product
	}

	_data struct {
		Import   *xml.КоммерческаяИнформация
		groups   map[string]*_group   // *xml.Группа
		products map[string]*_product // *xml.Товар
		Offers   *xml.КоммерческаяИнформация
		offers   map[string]*_offer // *xml.Предложение
	}
)

var (
	conf struct {
		Import string `envconfig:"default=import.xml"`
		Offers string `envconfig:"default=offers.xml"`
	}
)

func (d *_data) addGroups(groups *[]xml.Группа, path []string) {
	for idx := range *groups {
		g := &_group{
			Группа:   &(*groups)[idx],
			path:     append(path, (*groups)[idx].Наименование),
			products: make(map[string]*_product),
		}

		d.groups[g.Ид] = g
		if g.Группы != nil {
			d.addGroups(g.Группы, g.path)
		}
	}
}

func (d *_data) addProducts(products *[]xml.Товар) {
	for idx := range *products {
		p := &_product{
			Товар: &(*products)[idx],
		}

		d.products[p.Ид] = p
		for _, gid := range p.Группы {
			d.groups[gid].products[p.Ид] = p
		}
	}
}

func (d *_data) addOffers(offers *[]xml.Предложение) {
	for idx := range *offers {
		o := &_offer{
			Предложение: &(*offers)[idx],
		}

		d.offers[o.Ид] = o
	}
}

func (d *_data) init() {
	d.Import = xml.ReadMust(conf.Import)
	d.groups = make(map[string]*_group)
	d.addGroups(&d.Import.Классификатор.Группы, append([]string{}))
	d.products = make(map[string]*_product)
	d.addProducts(&d.Import.Каталог.Товары)

	d.Offers = xml.ReadMust(conf.Offers)
	d.offers = make(map[string]*_offer)
	d.addOffers(&d.Offers.ПакетПредложений.Предложения)

}

func main() {
	err := envconfig.InitWithPrefix(&conf, "CML2CSV")
	if err != nil {
		log.Fatal(err)
	}

	var data _data
	data.init()

	log.Printf("groups: %d goods: %d offers: %d",
		len(data.groups), len(data.products), len(data.offers))

	// CSV header
	// fmt.Printf("\n")

	// sku, path, name, ... offers
	for _, g := range data.groups {
		if len(g.products) == 0 {
			continue
		}

		for _, p := range g.products {
			o, found := data.offers[p.Ид]
			if !found {
				continue
			}

			fmt.Printf("%s; %s; %s; %s\n",
				p.Артикул,
				strings.Join(g.path, ">"),
				p.Наименование,
				o.Количество,
			)
		}
	}
}
