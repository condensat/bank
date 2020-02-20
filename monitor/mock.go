package monitor

const (
	jsonReportMock = `{
    "node1": {
      "coins": {
        "BTC": {
          "branch": "master",
          "commit": "36f42e1bf43f2c9f3b4642814051cedf66f05a5e",
          "repo": "https://github.com/bitcoin/bitcoin.git",
          "running": false
        },
        "CON": {
          "bestblockhash": "0376e26c97518160c28e4783716e15251ca1e6546278ef1690ad60d29c8c6e33",
          "blocks": 128,
          "branch": "master",
          "commit": "886280c5ffe087ad270ee157b7ecbb8f2890863c",
          "headers": 128,
          "repo": "https://github.com/KomodoPlatform/komodo.git",
          "running": true
        },
        "ELEMENTS": {
          "branch": null,
          "commit": null,
          "repo": null,
          "running": false
        },
        "KMD": {
          "branch": "master",
          "commit": "886280c5ffe087ad270ee157b7ecbb8f2890863c",
          "repo": "https://github.com/KomodoPlatform/komodo.git",
          "running": false
        }
      },
      "general": {
        "load": {
          "1": "0.00",
          "5": "0.00",
          "15": "0.00"
        },
        "now": "2020-02-20 07:27:06",
        "uptime": {
          "day": 2,
          "hour": 7,
          "minutes": 50,
          "second": 34
        }
      }
    },
    "node2": {
      "coins": {
        "BTC": {
          "branch": "master",
          "commit": "36f42e1bf43f2c9f3b4642814051cedf66f05a5e",
          "repo": "https://github.com/bitcoin/bitcoin.git",
          "running": false
        },
        "CON": {
          "bestblockhash": "0376e26c97518160c28e4783716e15251ca1e6546278ef1690ad60d29c8c6e33",
          "blocks": 128,
          "branch": "master",
          "commit": "886280c5ffe087ad270ee157b7ecbb8f2890863c",
          "headers": 128,
          "repo": "https://github.com/KomodoPlatform/komodo.git",
          "running": true
        },
        "ELEMENTS": {
          "branch": null,
          "commit": null,
          "repo": null,
          "running": false
        },
        "KMD": {
          "branch": "master",
          "commit": "886280c5ffe087ad270ee157b7ecbb8f2890863c",
          "repo": "https://github.com/KomodoPlatform/komodo.git",
          "running": false
        }
      },
      "general": {
        "load": {
          "1": "0.01",
          "5": "0.00",
          "15": "0.00"
        },
        "now": "2020-02-20 06:50:54",
        "uptime": {
          "day": 2,
          "hour": 7,
          "minutes": 14,
          "second": 23
        }
      }
    },
    "node4": {
      "coins": {
        "BTC": {
          "branch": "master",
          "commit": "36f42e1bf43f2c9f3b4642814051cedf66f05a5e",
          "repo": "https://github.com/bitcoin/bitcoin.git",
          "running": false
        },
        "CON": {
          "bestblockhash": "0376e26c97518160c28e4783716e15251ca1e6546278ef1690ad60d29c8c6e33",
          "blocks": 128,
          "branch": "master",
          "commit": "886280c5ffe087ad270ee157b7ecbb8f2890863c",
          "headers": 128,
          "repo": "https://github.com/KomodoPlatform/komodo.git",
          "running": true
        },
        "ELEMENTS": {
          "branch": null,
          "commit": null,
          "repo": null,
          "running": false
        },
        "KMD": {
          "branch": "master",
          "commit": "886280c5ffe087ad270ee157b7ecbb8f2890863c",
          "repo": "https://github.com/KomodoPlatform/komodo.git",
          "running": false
        }
      },
      "general": {
        "load": {
          "1": "0.01",
          "5": "0.00",
          "15": "0.00"
        },
        "now": "2020-02-20 06:50:54",
        "uptime": {
          "day": 2,
          "hour": 7,
          "minutes": 14,
          "second": 23
        }
      }
    }
  }
`
)
