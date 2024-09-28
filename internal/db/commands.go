package db

func getKey(key string) (string, error) {
	val, err := rdb.Get(ctx, key).Result()
	return val, err
}

func setKey(key, value string) error {
	err := rdb.Set(ctx, key, value, 0).Err()
	return err
}

func deleteKeys(keys []string) error {
	err := rdb.Del(ctx, keys...).Err()
	return err
}
