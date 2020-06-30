Update confluence wiki page use markdown

# Installation

```bash
$ go get git.yunion.io/et/md2cflc
```

# Usage

```bash
# convert markdown to confluence wiki markup content
$ md2cflc ./README.md

# update your confluence wiki page by pageid
$ md2cflc -wiki https://wiki.xxx.com -u username -p passwd -pageid 12345 ./README.md 
```

# FAQ

## How to get page id?

You can find the page id when you edit it.

![Find pageId](./img/pageId.png)
