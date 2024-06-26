package database

import (
	"cmp"
	"errors"
	"fmt"
	helperFuncs "pkg/helper"
	"slices"

	badger "github.com/dgraph-io/badger/v4"
)

func (bs *BadgerDBStore) GetDay(date Date) (DailyStat, error) {

	if day, _ := ParseKey(date); day.After(formattedToDay()) {
		return ZeroValueDailyStat, ErrFutureDay
	}

	byteData, err := bs.Get(dbDayKey(date))
	errKeyNotFound := errors.Is(err, badger.ErrKeyNotFound)

	if err != nil && !errKeyNotFound {
		return ZeroValueDailyStat, err
	}

	if errKeyNotFound {
		return bs.getDailyAppStat(date)
	}

	dayStat, err := helperFuncs.Decode[DailyStat](byteData)
	if err != nil {
		return ZeroValueDailyStat, err
	}
	return dayStat, nil
}

func (bs *BadgerDBStore) getDailyAppStat(day Date) (DailyStat, error) {
	var (
		result       DailyStat
		dayTotalData Stats
		arr          = make([]AppStat, 0, 20)
	)

	err := bs.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = true
		opts.PrefetchSize = 100
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := dbAppPrefix
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			err := it.Item().Value(func(v []byte) error {

				var (
					app            AppInfo
					appStatArrData AppStat
					err            error
				)

				if app, err = helperFuncs.Decode[AppInfo](v); err != nil {
					return err
				}

				thatDayStat := app.ScreenStat[day]

				// EachApp []appstat
				appStatArrData.AppName = app.AppName
				appStatArrData.Usage.Active = thatDayStat.Active
				appStatArrData.Usage.Inactive = thatDayStat.Inactive
				appStatArrData.Usage.Open = thatDayStat.Open
				arr = append(arr, appStatArrData)

				// DayTotall stats
				dayTotalData.Active += thatDayStat.Active
				dayTotalData.Inactive += thatDayStat.Inactive
				dayTotalData.Open += thatDayStat.Open

				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return ZeroValueDailyStat, err
	}

	slices.SortFunc(arr, func(a, b AppStat) int {
		return cmp.Compare(b.Usage.Active, a.Usage.Active)
	})

	result.DayTotal.Active = dayTotalData.Active
	result.DayTotal.Inactive = dayTotalData.Inactive
	result.DayTotal.Open = dayTotalData.Open
	result.EachApp = arr

	if day != Date(formattedToDay().Format(timeFormat)) {
		byteData, _ := helperFuncs.Encode(result)
		err := bs.setNewEntryToDB(dbDayKey(day), byteData)
		if err != nil {
			fmt.Println("ERROR WRITING NEW DAY ENTRY", day, "ERROR IS:", err)
		} else {
			fmt.Println("WRITING NEW DAY ENTRY", day)
		}
	}

	return result, nil
}
