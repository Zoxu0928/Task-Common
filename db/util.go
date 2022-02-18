package db

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var (
	dbFieldRegex = regexp.MustCompile(`([\w|_.]+)([_][\d]+)`)
)

type Pages struct {
	PageNumber int64    `json:"pageNumber"`
	PageSize   int64    `json:"pageSize"`
	Order      []string `json:"order"`
	Sort       string   `json:"sort"`
}

// func NewPage(order, sort string, descOffset, descLimit int64) *Pages {

// func (pg *Pages) GeneratePageSql() string {

// type OrderSort struct {

// type PageSize struct {

// func NewPageSize(OrderBy []*OrderSort, descOffset, descLimit int) *PageSize {

// func (pg *PageSize) GeneratePageSizeSql() string {

// func GenerateQuerySql(selectedFields, table, whereSql string) string {

// func GenerateDeleteSql(table, whereSql string) string {

// func UpdateSqlWhereIdGenerate(table string, updateFields []string) string {

// func UpdateSqlGenerate(table string, updateFields, whereFields []string) string {

// func UpdateSqlGenerateByValue(table string, updateFields, whereFields []string, whereParam interface{}) (string, error) {

// func WhereSqlGenerate(whereFields []string, param interface{}) (string, error) {

// func UpdateResourceSqlGenerate(table string, updateFields, whereFields []string) string {

func parasDBField(field string) (string, bool) {

	if dbFieldRegex.MatchString(field) {
		return dbFieldRegex.FindStringSubmatch(field)[1], true
	}

	return field, false
}

// func BatchFields(dbField string, batchFields []string, condition map[string]interface{}) {

// func BatchIdFields(dbField string, batchFields []int64, condition map[string]interface{}) {

func BatchFieldsInterface(dbField string, batchFields []interface{}, condition map[string]interface{}) {
	if len(batchFields) == 0 {
		return
	}
	for index, id := range batchFields {
		condition[dbField+"_"+strconv.Itoa(index)] = id
	}
}

func GenerateFilterSql(condition map[string]interface{}, operaters map[string]string) string {
	querySql := " "
	first := 0
	// 只支持一个批量字段
	orSqls := make(map[string]string)
	for k, _ := range condition {
		f, ok := parasDBField(k)
		if ok {
			var orSql string
			if operater, ok := operaters[f]; ok {
				orSql = fmt.Sprintf("%s %s @%s or ", f, operater, k)
			} else {
				orSql = fmt.Sprintf("%s = @%s or ", f, k)
			}
			if _, ok := orSqls[f]; ok {
				orSqls[f] = orSqls[f] + orSql
				continue
			}
			orSqls[f] = orSql
			continue
		}
		first++
		if first > 1 {
			querySql += " and "
		}

		if operater, ok := operaters[k]; ok {
			querySql += fmt.Sprintf("%s %s @%s", k, operater, k)
		} else {
			querySql += fmt.Sprintf("%s = @%s", k, k)
		}
	}

	if len(orSqls) > 0 {
		i := 0
		for _, v := range orSqls {
			if i == 0 && querySql == " " {
				querySql += " (" + v[:len(v)-3] + ")"
			} else {
				querySql += " and (" + v[:len(v)-3] + ")"
			}
			i++
		}
	}

	return strings.Trim(querySql, " ")
}

// func GenerateFilterSqlWithPrefix(condition map[string]interface{}, operaters map[string]string, prefix string) string {

// func GenerateBatchQuerySQL(base string, condition map[string]interface{}) string {

// func GenerateFilterDurationDeleteSql(condition map[string]interface{}, operaters map[string]string, needJoin bool) string {

type MysqlInstance struct {
	Conf   *DatabaseConf
	DBOpts *gorm.Config
}

func (ins MysqlInstance) NewMysql() (*gorm.DB, error) {
	return ins.NewMysqlByDBOpts(ins.DBOpts)
}

func (ins MysqlInstance) NewMysqlByDBOpts(opts *gorm.Config) (*gorm.DB, error) {
	if opts == nil {
		opts = &gorm.Config{
			NamingStrategy: schema.NamingStrategy{
				// 使用单数表名，启用该选项后，`Task` 表将是 `task` 而不是 `tasks`
				SingularTable: true,
			},
			// 在完成初始化后，GORM 会自动 ping 数据库以检查数据库的可用性，若要禁用该特性，可将其设置为 true
			DisableAutomaticPing: true,
		}
	}
	ins.DBOpts = opts
	return ins.newMysqlInstance()
}

func (ins MysqlInstance) newMysqlInstance() (*gorm.DB, error) {
	// data source name
	// transaction_isolation 是 tx_isolation 的别名（MySQL@>=5.7.20）
	// tx_isolation（MySQL@<8.0、MariaDB）
	dsnTpl := "%s:%s@tcp(%s:%d)/%s?autocommit=true&parseTime=true&timeout=%dms&loc=Asia%%2FShanghai&transaction_isolation='READ-COMMITTED'"
	dsn := fmt.Sprintf(
		dsnTpl,
		ins.Conf.User,
		ins.Conf.Password,
		ins.Conf.Ip,
		ins.Conf.Port,
		ins.Conf.DB,
		ins.Conf.Timeout)

	db, err := gorm.Open(mysql.Open(dsn), ins.DBOpts)
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(ins.Conf.MaxConnection)
	sqlDB.SetMaxIdleConns(ins.Conf.MaxIdleConnection)
	sqlDB.SetConnMaxLifetime(time.Second * time.Duration(ins.Conf.MaxLifetime))

	err = sqlDB.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}

// func Fieldjoin(fields string) (r string) {
