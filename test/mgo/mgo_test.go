package main

import (
	"testing"

	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/db"
	"github.com/davyxu/cellnet/test"
	"github.com/davyxu/golog"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var log *golog.Logger = golog.New("test")

var signal *test.SignalTester

type char struct {
	Name string
	HP   int
}

func update(db *db.MongoDriver, evq cellnet.EventQueue) {
	log.Debugln("update")

	db.Execute(func(ses *mgo.Session) {

		col := ses.DB("").C("test")

		col.Update(bson.M{"name": "davy"}, &char{Name: "davy", HP: 1})

		evq.PostData(func() {
			signal.Done(2)
		})
	})

}

func rundb() {
	pipe := cellnet.NewEventPipe()

	evq := pipe.AddQueue()

	pipe.Start()

	mdb := db.NewMongoDriver()

	var err error

	err = mdb.Start(&db.Config{
		URL:       "127.0.0.1:27017/test",
		ConnCount: 1,
	})

	if err != nil {
		signal.Fail()
		return
	}

	mdb.Execute(func(ses *mgo.Session) {

		col := ses.DB("").C("test")

		var c char

		err := col.Find(bson.M{"name": "davy"}).One(&c)

		evq.PostData(func() {

			if err == mgo.ErrNotFound {

				mdb.Execute(func(ses *mgo.Session) {

					col := ses.DB("").C("test")

					log.Debugln("insert new")

					col.Insert(&char{Name: "davy", HP: 10})
					col.Insert(&char{Name: "zerg", HP: 90})

					evq.PostData(func() {

						signal.Done(1)

						update(mdb, evq)
					})

				})

			} else {

				log.Debugln("exist")

				log.Debugln(c)

				signal.Done(1)
				update(mdb, evq)
			}
		})

	})

	signal.WaitAndExpect(1, "find failed")
	signal.WaitAndExpect(2, "update failed")

}

func TestMongoDB(t *testing.T) {

	signal = test.NewSignalTester(t)

	rundb()
}
