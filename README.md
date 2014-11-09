# gtig - go twitter irc gateway

pure go irc server + twitter listener

gtigはpure goでかかれたなんとなくircサーバーで、twitterのirc gatewayのサンプルがついています。
とにかく楽にカスタムできるようにevent dispatchで挙動をゴリゴリ書き換えれます。

IRCサーバーとしては未実装なものが多すぎるし適当すぎるのですが、とりあえずIRCでアクセスできるようにしたい
みたいなときに便利かもしれませんね

# Dependencies

```
github.com/garyburd/go-oauth
gopkg.in/yaml.v2
```

# つかいかた

twitter.goにoauth tokenとsecretがあるんで自分のaccess tokenをなんとかして拾ってきてください。
config.ymlを下記の通りつくっといてgo run main.goすればおｋ

```
oauth:
  token: "ひろったaccess token"
  secret:  "ひろったaccess token seret"
```

# License

MIT License