# s3-obfuscation


## Desc
I create this because i don't like my colleague can peek my s3 bucket.

this project does not have any security concern so **don't use with your important data.**

## How to use
```bash
cp .env.template .env
nano .env
export $(cat .env | xargs)
go run cmd/main.go -bucket test -passphrase SuperSecretAgent#2! init
go run cmd/main.go -bucket test -passphrase SuperSecretAgent#21 put test/main.go
go run cmd/main.go -bucket test -passphrase SuperSecretAgent#21 list 
```

## Note
This tools will save your s3 information in "index" file on your bucket, don't delete the index file otherwise your obj cannot be decrypted.

In prevent to lost your index file you can save it on your local or you can use `list -s` command and it's will save your index on decrypted state

## Roadmap (?)
I'm not sure about this one since this only have fun project but i feel i can make this project more look professional.

The only problem right now is `index` file [SPOF](https://en.wikipedia.org/wiki/Single_point_of_failure), I was thinking to avoid this SPOF is to distribute every index file into each directory

Right now the structure is like this:
```bash
╭─[humanz-403] as humanz in /mnt/Data/P
╰──➤ s3cmd --config=s3config ls s3://test
                          DIR  s3://test/5fa62ae6176f3746142503a6ebe96cb3/
                          DIR  s3://test/be383898a86f10db3a7a123ec4afedc9/
2025-03-12 06:58          439  s3://test/index
```

And what it's should be is like this:
```bash
╭─[humanz-403] as humanz in /mnt/Data/P
╰──➤ s3cmd --config=s3config ls s3://test --recursive
2025-03-11 23:36       439    s3://test/5fa62ae6176f3746142503a6ebe96cb3/index
2025-03-11 23:36       439    s3://test/5fa62ae6176f3746142503a6ebe96cb3/82048665ff37923fe93dcd2ca42115dc/index
2025-03-11 23:36       439    s3://test/5fa62ae6176f3746142503a6ebe96cb3/82048665ff37923fe93dcd2ca42115dc/9376f1a0002a78696f71ade594a287a0/index
2025-03-11 23:36       439    s3://test/5fa62ae6176f3746142503a6ebe96cb3/82048665ff37923fe93dcd2ca42115dc/9376f1a0002a78696f71ade594a287a0/556933f099ccf5c2ac1e2179faa0903b
2025-03-12 06:58          439  s3://test/index
```


