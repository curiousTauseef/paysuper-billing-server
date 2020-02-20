package database

import (
	"fmt"
	"github.com/go-redis/redis"
	_ "github.com/golang-migrate/migrate/v4/database/mongodb"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/paysuper/paysuper-billing-server/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"testing"
	"time"
)

type CacheTestSuite struct {
	suite.Suite
	redis redis.Cmdable
	cache CacheInterface
	log   *zap.Logger
}

type Test struct {
	Message string `json:"message"`
	Value   int32  `json:"value"`
	Hidden  string `json:"-"`
	Child   *Test  `json:"child"`
}

func Test_Cache(t *testing.T) {
	suite.Run(t, new(CacheTestSuite))
}

func (suite *CacheTestSuite) SetupTest() {
	cfg, err := config.NewConfig()
	assert.NoError(suite.T(), err, "Config load failed")

	suite.redis = NewRedis(
		&redis.Options{
			Addr:     cfg.RedisHost,
			Password: cfg.RedisPassword,
		},
	)
	suite.cache, err = NewCacheRedis(suite.redis, "cache")
	assert.NoError(suite.T(), err)
}

func (suite *CacheTestSuite) TearDownTest() {
	suite.cache.FlushAll()
}

func (suite *CacheTestSuite) TestCache_CleanOldestVersion_NoOldestVersions() {
	err := suite.cache.CleanOldestVersion()
	assert.NoError(suite.T(), err)
}

func (suite *CacheTestSuite) TestCache_CleanOldestVersion_ReturnTrue() {
	for i := 0; i <= versionLimit; i++ {
		_, err := NewCacheRedis(suite.redis, fmt.Sprintf("cache%d", i))
		assert.NoError(suite.T(), err)
	}
	err := suite.cache.CleanOldestVersion()
	assert.NoError(suite.T(), err)
}

func (suite *CacheTestSuite) TestCache_CleanOldestVersion_SuccessfullyDeletedKeys() {
	oldestCache, _ := NewCacheRedis(suite.redis, "cache_old")
	_, err := NewCacheRedis(suite.redis, "cache_new1")
	assert.NoError(suite.T(), err)
	_, err = NewCacheRedis(suite.redis, "cache_new2")
	assert.NoError(suite.T(), err)

	var obj = &Test{
		Message: "message",
		Value:   1,
		Hidden:  "hidden",
	}
	_ = oldestCache.Set("test", obj, 0)

	var obj1 = &Test{}
	_ = oldestCache.Get("test", obj1)
	assert.NotEmpty(suite.T(), obj1)

	err = suite.cache.CleanOldestVersion()
	assert.NoError(suite.T(), err)

	time.Sleep(2 * time.Second)

	var obj2 = &Test{}
	_ = oldestCache.Get("test", obj2)
	assert.Empty(suite.T(), obj2)
}

func (suite *CacheTestSuite) TestCache_WriteRead_JsonTag() {
	cache, err := NewCacheRedis(suite.redis, "cache")
	assert.NoError(suite.T(), err)
	var obj = &Test{
		Message: "message",
		Value:   1,
		Hidden:  "hidden",
	}
	err = cache.Set("test", obj, 0)
	assert.NoError(suite.T(), err)
	var obj2 = &Test{}
	err = cache.Get("test", obj2)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), obj2)
	assert.EqualValues(suite.T(), obj.Message, obj2.Message)
	assert.EqualValues(suite.T(), obj.Value, obj2.Value)
	assert.EqualValues(suite.T(), obj.Hidden, obj2.Hidden)
}

func (suite *CacheTestSuite) TestCache_WriteRead_Nesting() {
	cache, err := NewCacheRedis(suite.redis, "cache")
	assert.NoError(suite.T(), err)
	var obj = &Test{
		Child: &Test{
			Message: "message",
		},
	}
	err = cache.Set("test", obj, 0)
	assert.NoError(suite.T(), err)
	var obj2 = &Test{}
	err = cache.Get("test", obj2)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), obj2)
	assert.NotEmpty(suite.T(), obj2.Child)
	assert.EqualValues(suite.T(), obj.Child.Message, obj2.Child.Message)
}
