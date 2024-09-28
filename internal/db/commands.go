package db

func (db *DiceDB) getKey(key string) (string, error) {
	val, err := db.Client.Get(db.Ctx, key).Result()
	return val, err
}

func (db *DiceDB) setKey(key, value string) error {
	err := db.Client.Set(db.Ctx, key, value, 0).Err()
	return err
}

func (db *DiceDB) deleteKeys(keys []string) error {
	err := db.Client.Del(db.Ctx, keys...).Err()
	return err
}
