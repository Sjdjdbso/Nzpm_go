package ext_utils

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DbManager struct {
	client *mongo.Client
	db     *mongo.Database
	botID  string
	mu     sync.RWMutex
}

var DB *DbManager

// InitDB menginisialisasi koneksi singleton ke MongoDB Atlas (persis wzv3 db_handler.py)
func InitDB(databaseURL string, botID int64) (*DbManager, error) {
	if databaseURL == "" {
		log.Println("[INFO] DATABASE_URL kosong. Menggunakan penyimpanan lokal (In-Memory & JSON).")
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(databaseURL))
	if err != nil {
		return nil, fmt.Errorf("gagal menghubungkan ke MongoDB: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("gagal ping MongoDB: %w", err)
	}

	db := client.Database("wzmlx")
	bIDStr := fmt.Sprintf("%d", botID)
	if botID == 0 {
		bIDStr = "default"
	}

	DB = &DbManager{
		client: client,
		db:     db,
		botID:  bIDStr,
	}

	log.Printf("[MONGODB] Berhasil terhubung ke MongoDB Atlas (Database: wzmlx, BotID: %s)", bIDStr)

	// Sinkronisasi data saat startup (persis wzv3 db_load)
	go DB.DbLoad()

	return DB, nil
}

// DbLoad memuat semua data pengguna, custom thumbnails, dan pengaturan dari MongoDB
func (m *DbManager) DbLoad() {
	if m == nil || m.db == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 1. Muat data pengguna (users.<bot_id>)
	usersCol := m.db.Collection("users." + m.botID)
	cursor, err := usersCol.Find(ctx, bson.M{})
	if err == nil {
		defer cursor.Close(ctx)
		count := 0
		_ = os.MkdirAll("thumbnails", 0755)
		_ = os.MkdirAll("rclone", 0755)

		for cursor.Next(ctx) {
			var doc bson.M
			if err := cursor.Decode(&doc); err != nil {
				continue
			}

			uidRaw, ok := doc["_id"]
			if !ok {
				continue
			}

			var uid int64
			switch v := uidRaw.(type) {
			case int64:
				uid = v
			case int32:
				uid = int64(v)
			case float64:
				uid = int64(v)
			}

			uCfg := UserStore.Get(uid)

			if cap, ok := doc["caption"].(string); ok {
				uCfg.CustomCaption = cap
			}
			if pfx, ok := doc["prefix"].(string); ok {
				uCfg.LeechPrefix = pfx
			}
			if sfx, ok := doc["suffix"].(string); ok {
				uCfg.LeechSuffix = sfx
			}
			if pdAPI, ok := doc["pixeldrain_api"].(string); ok {
				uCfg.PixeldrainAPI = pdAPI
			}
			// Parse format ddl_servers dari wzv3 jika ada
			if ddlServers, ok := doc["ddl_servers"].(bson.M); ok {
				if pdData, ok := ddlServers["pixeldrain"].(bson.A); ok && len(pdData) > 1 {
					if key, ok := pdData[1].(string); ok && key != "" {
						uCfg.PixeldrainAPI = key
					}
				}
			}

			// Simpan thumbnail binary ke file lokal jika ada di MongoDB
			if thumbBin, ok := doc["thumb"].([]byte); ok && len(thumbBin) > 0 {
				thumbPath := filepath.Join("thumbnails", fmt.Sprintf("%d.jpg", uid))
				_ = os.WriteFile(thumbPath, thumbBin, 0644)
				uCfg.HasThumbnail = true
			}

			// Simpan per-user rclone.conf binary jika ada di MongoDB
			if rcloneBin, ok := doc["rclone"].([]byte); ok && len(rcloneBin) > 0 {
				rcPath := filepath.Join("rclone", fmt.Sprintf("%d.conf", uid))
				_ = os.WriteFile(rcPath, rcloneBin, 0644)
			}

			UserStore.Lock()
			UserStore.Users[uid] = uCfg
			UserStore.Unlock()
			count++
		}
		if count > 0 {
			log.Printf("[MONGODB] Berhasil memuat %d data pengaturan pengguna dari Database", count)
		}
	}
}

// UpdateUserData menyimpan pengaturan pengguna ke MongoDB
func (m *DbManager) UpdateUserData(userID int64, cfg *UserConfig) error {
	if m == nil || m.db == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	col := m.db.Collection("users." + m.botID)
	update := bson.M{
		"$set": bson.M{
			"caption":        cfg.CustomCaption,
			"prefix":         cfg.LeechPrefix,
			"suffix":         cfg.LeechSuffix,
			"pixeldrain_api": cfg.PixeldrainAPI,
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err := col.UpdateByID(ctx, userID, update, opts)
	return err
}

// UpdateUserThumb menyimpan file thumbnail binary ke MongoDB
func (m *DbManager) UpdateUserThumb(userID int64, data []byte) error {
	if m == nil || m.db == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	col := m.db.Collection("users." + m.botID)
	update := bson.M{
		"$set": bson.M{
			"thumb": data,
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err := col.UpdateByID(ctx, userID, update, opts)
	return err
}

// DeleteUserThumb menghapus thumbnail dari MongoDB
func (m *DbManager) DeleteUserThumb(userID int64) error {
	if m == nil || m.db == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	col := m.db.Collection("users." + m.botID)
	update := bson.M{
		"$unset": bson.M{
			"thumb": "",
		},
	}
	_, err := col.UpdateByID(ctx, userID, update)
	return err
}

// UpdateUserRclone menyimpan file rclone.conf milik user ke MongoDB
func (m *DbManager) UpdateUserRclone(userID int64, data []byte) error {
	if m == nil || m.db == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	col := m.db.Collection("users." + m.botID)
	update := bson.M{
		"$set": bson.M{
			"rclone": data,
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err := col.UpdateByID(ctx, userID, update, opts)
	return err
}

// DeleteUserRclone menghapus file rclone user dari MongoDB
func (m *DbManager) DeleteUserRclone(userID int64) error {
	if m == nil || m.db == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	col := m.db.Collection("users." + m.botID)
	update := bson.M{
		"$unset": bson.M{
			"rclone": "",
		},
	}
	_, err := col.UpdateByID(ctx, userID, update)
	return err
}

// AddPMUser menambahkan ID pengguna yang membuka chat PM ke pm_users.<bot_id>
func (m *DbManager) AddPMUser(userID int64) error {
	if m == nil || m.db == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	col := m.db.Collection("pm_users." + m.botID)
	opts := options.Update().SetUpsert(true)
	_, err := col.UpdateByID(ctx, userID, bson.M{"$set": bson.M{"_id": userID}}, opts)
	return err
}

// GetAllPMUsers mengambil seluruh ID pengguna PM untuk broadcast
func (m *DbManager) GetAllPMUsers() ([]int64, error) {
	if m == nil || m.db == nil {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	col := m.db.Collection("pm_users." + m.botID)
	cursor, err := col.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var uids []int64
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err == nil {
			if idRaw, ok := doc["_id"]; ok {
				switch v := idRaw.(type) {
				case int64:
					uids = append(uids, v)
				case int32:
					uids = append(uids, int64(v))
				case float64:
					uids = append(uids, int64(v))
				}
			}
		}
	}
	return uids, nil
}
