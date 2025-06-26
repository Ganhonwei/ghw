module goserver

go 1.23.4

replace github.com/AsynkronIT/protoactor-go => ./third_party/github.com/AsynkronIT/protoactor-go

replace github.com/gogo/protobuf => ./third_party/github.com/gogo/protobuf

replace github.com/op/go-logging => ./third_party/github.com/op/go-logging

replace github.com/360EntSecGroup-Skylar/excelize/v2 => github.com/xuri/excelize/v2 v2.8.0

replace github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.9.3

require (
	github.com/360EntSecGroup-Skylar/excelize v1.4.1
	github.com/360EntSecGroup-Skylar/excelize/v2 v2.9.0
	github.com/AsynkronIT/goconsole v0.0.0-20160504192649-bfa12eebf716
	github.com/AsynkronIT/protoactor-go v0.0.0-20240822202345-3c0e61ca19c9
	github.com/ClickHouse/clickhouse-go/v2 v2.30.0
	github.com/EDDYCJY/fake-useragent v0.2.0
	github.com/astaxie/beego v1.12.3
	github.com/aws/aws-sdk-go v1.43.16
	github.com/beego/i18n v0.0.0-20161101132742-e9308947f407
	github.com/bwmarrin/snowflake v0.3.0
	github.com/bytedance/sonic v1.12.5
	github.com/dhushon/decxls v0.0.0-20201203214035-0cd3f5389616
	github.com/emirpasic/gods v1.18.1
	github.com/globalsign/mgo v0.0.0-20181015135952-eeefdecb41b8
	github.com/go-redsync/redsync/v4 v4.13.0
	github.com/go-resty/resty/v2 v2.16.2
	github.com/go-telegram-bot-api/telegram-bot-api v4.6.4+incompatible
	github.com/gogo/protobuf v1.3.2
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/golang/protobuf v1.5.4
	github.com/google/generative-ai-go v0.19.0
	github.com/google/uuid v1.6.0
	github.com/gorilla/websocket v1.5.3
	github.com/hibiken/asynq v0.25.1
	github.com/json-iterator/go v1.1.12
	github.com/lab259/cors v0.2.0
	github.com/matoous/go-nanoid/v2 v2.1.0
	github.com/mattn/go-runewidth v0.0.16
	github.com/nats-io/nats.go v1.38.0
	github.com/nikoksr/notify v1.1.0
	github.com/nsqio/go-nsq v1.1.0
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/oschwald/geoip2-golang v1.11.0
	github.com/panjf2000/ants/v2 v2.11.1
	github.com/pkg/errors v0.9.1
	github.com/pkg/sftp v1.13.1
	github.com/redis/go-redis/v9 v9.7.0
	github.com/robfig/cron/v3 v3.0.1
	github.com/scylladb/termtables v0.0.0-20191203121021-c4c0b6d42ff4
	github.com/segmentio/kafka-go v0.4.47
	github.com/shopspring/decimal v1.4.0
	github.com/spf13/cobra v1.9.1
	github.com/valyala/fasthttp v1.58.0
	github.com/valyala/quicktemplate v1.8.0
	github.com/wizjin/weixin v0.0.0-20180905024723-e9f0f2024d5d
	go.uber.org/zap v1.27.0
	golang.org/x/crypto v0.31.0
	golang.org/x/sync v0.11.0
	golang.org/x/time v0.8.0
	google.golang.org/api v0.204.0
	google.golang.org/grpc v1.67.1
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
	gopkg.in/ini.v1 v1.67.0
	gopkg.in/mgo.v2 v2.0.0-20190816093944-a6b53ec6cb22
	gopkg.in/telebot.v3 v3.3.8
	gorm.io/driver/clickhouse v0.6.1
	gorm.io/gorm v1.25.12
)

require (
	cloud.google.com/go v0.116.0 // indirect
	cloud.google.com/go/ai v0.8.0 // indirect
	cloud.google.com/go/auth v0.10.0 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.5 // indirect
	cloud.google.com/go/compute/metadata v0.5.2 // indirect
	cloud.google.com/go/longrunning v0.6.2 // indirect
	github.com/ClickHouse/ch-go v0.61.5 // indirect
	github.com/PuerkitoBio/goquery v1.10.0 // indirect
	github.com/Unknwon/goconfig v1.0.0 // indirect
	github.com/Workiva/go-datastructures v1.1.5 // indirect
	github.com/andybalholm/brotli v1.1.1 // indirect
	github.com/andybalholm/cascadia v1.3.2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bytedance/sonic/loader v0.2.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cloudwego/base64x v0.1.4 // indirect
	github.com/cloudwego/iasm v0.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-faster/city v1.0.1 // indirect
	github.com/go-faster/errors v0.7.1 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/s2a-go v0.1.8 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.4 // indirect
	github.com/googleapis/gax-go/v2 v2.13.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-version v1.6.0 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/klauspost/cpuid/v2 v2.0.9 // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/nats-io/nkeys v0.4.9 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/orcaman/concurrent-map v1.0.0 // indirect
	github.com/oschwald/maxminddb-golang v1.13.0 // indirect
	github.com/paulmach/orb v0.11.1 // indirect
	github.com/pierrec/lz4/v4 v4.1.21 // indirect
	github.com/prometheus/client_golang v1.14.0 // indirect
	github.com/prometheus/client_model v0.4.0 // indirect
	github.com/prometheus/common v0.37.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/richardlehane/mscfb v1.0.4 // indirect
	github.com/richardlehane/msoleps v1.0.3 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	github.com/serialx/hashring v0.0.0-20200727003509-22c0c7ab6b1b // indirect
	github.com/shiena/ansicolor v0.0.0-20151119151921-a422bbe96644 // indirect
	github.com/spf13/cast v1.7.0 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	github.com/technoweenie/multipartstreamer v1.0.1 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/xuri/efp v0.0.0-20230802181842-ad255f2331ca // indirect
	github.com/xuri/nfp v0.0.0-20230819163627-dc951e3ffe1a // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.56.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.56.0 // indirect
	go.opentelemetry.io/otel v1.31.0 // indirect
	go.opentelemetry.io/otel/metric v1.31.0 // indirect
	go.opentelemetry.io/otel/trace v1.31.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/arch v0.0.0-20210923205945-b76863e36670 // indirect
	golang.org/x/net v0.31.0 // indirect
	golang.org/x/oauth2 v0.23.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20241021214115-324edc3d5d38 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241021214115-324edc3d5d38 // indirect
	google.golang.org/protobuf v1.35.2 // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
