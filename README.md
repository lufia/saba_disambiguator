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
% make import
```

## 分類器の学習
データの整備が完了したので、教師データを用いて以下のコマンドで分類器(平均化パーセプトロン)を学習させます。

```
% make learn
```

学習の完了後、`functions/saba_disambiguator/build/model.bin`というファイルが自動生成されているはずです。

# AWS Lambdaで動かす
## 設定ファイル
動かす前に設定が必要です。設定は`functions/saba_disambiguator/build/config.yml`に書きます。`functions/saba_disambiguator/build/config_sample.yml`にサンプルがあるので、それを参考にするとよいでしょう。secretキーなどはリポジトリで管理したくない情報なので、[AWS Systems Manager パラメータストア](https://docs.aws.amazon.com/ja_jp/systems-manager/latest/userguide/systems-manager-parameter-store.html)で管理します

## Deploy
AWS Lambdaへのdeployは[SAM](https://aws.amazon.com/jp/serverless/sam/)を使います。以下のコマンドでdeployできます。

```
% make sam-package sam-deploy
```

## CloudWatchイベントを用いてスケジューリングする
SAMでdeployをすると自動的にCloudWatchイベントでスケジュールされます。間隔を変更したい場合はスケジューリングをoffにしたい場合は`template.yml`を変更してから再deployしましょう。
