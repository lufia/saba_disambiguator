# saba_disambiguator
Twitterの`mackerel`の検索結果から[mackerel.io](https://mackerel.io)に関連するものだけをSlackに通知するスクリプトです。AWS Lambda上で定期実行して動かします。

# Overview
- AWS Lambda上で直近の`mackerel`に関連するtweetを定期的に収集
- tweetのjsonから特徴量を生成し、分類器にかけます
- 正例であると判定されたtweetは所望のSlackのチャンネルに通知されます

# 下準備
## 教師ラベルの整備
教師あり学習を行なうため、教師ラベルを必要とします。`data/pos.txt`と`data/neg.txt`に正例と負例それぞれのtweetのパーマリンクを1行1tweetで書いていきます。例えば、`data/pos.txt`に以下のように書いていきます。

```
https://twitter.com/syou6162/status/931754069806297089
https://twitter.com/mackerelio_jp/status/931369140534747137
```

## JSONファイルの収集
教師ラベルが整備できたら各tweetに対応するJSONファイルを収集します。これでデータの整備は完了です。

```
% cat data/pos.txt | go run import_json.go > pos.json
% cat data/neg.txt | go run import_json.go > neg.json
```

## 分類器の学習
データの整備が完了したので、教師データを用いて以下のコマンドで分類器(平均化パーセプトロン)を学習させます。

```
% go run train_perceptron.go pos.json neg.json
```

学習の完了後、`model/model.bin`というファイルが自動生成されているはずです。AWS Lambdaにシングルバイナリで転送したい都合上、モデルファイルもgoのプログラムである必要があります。[go-bindata](https://github.com/jteeuwen/go-bindata)を用いて変換します。変換が正しく行なわれていれば、`lib/model.go`というファイルが生成されているはずです。

```
% go-bindata -pkg=sabadisambiguator -o=lib/model.go model/
```

