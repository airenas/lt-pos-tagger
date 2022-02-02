# LT Part of Speech tagging service

![Go](https://github.com/airenas/lt-pos-tagger/workflows/Go/badge.svg) [![Coverage Status](https://coveralls.io/repos/github/airenas/lt-pos-tagger/badge.svg?branch=main)](https://coveralls.io/github/airenas/lt-pos-tagger?branch=main) [![Go Report Card](https://goreportcard.com/badge/github.com/airenas/lt-pos-tagger)](https://goreportcard.com/report/github.com/airenas/lt-pos-tagger) ![CodeQL](https://github.com/airenas/lt-pos-tagger/workflows/CodeQL/badge.svg) [![Load Tests](https://github.com/airenas/lt-pos-tagger/actions/workflows/load.yml/badge.svg)](https://github.com/airenas/lt-pos-tagger/actions/workflows/load.yml) [![Integration Tests](https://github.com/airenas/lt-pos-tagger/actions/workflows/integration.yml/badge.svg)](https://github.com/airenas/lt-pos-tagger/actions/workflows/integration.yml)


Lithuanian Part of Speech Tagger - easy to use wrapper for lex and morph. The repository implements a service wrapper for *semantikadocker.vdu.lt/v2/morph* and *semantikadocker.vdu.lt/lex* services. These both services have quite complex API. This service makes the POS tag output simple to use and to understand. 

Also it fixes some issues with *lex* segmentation.

## Deploy

Deployment sample is prepared with *docker*: [example/docker-compose.yml](example/docker-compose.yml). You are on Linux? To start a service locally: 

```bash
   cd example 
   make up
```

That's it. You can start using the service:
```bash
   curl -X POST http://localhost:8092/tag -d 'Mama su kasa kasa smėlį.'
```

The output is expected to be the list of tagged words:

```json
[
  {
    "type": "WORD",
    "string": "Mama",
    "mi": "Ncfsnn-",
    "lemma": "mama"
  },
  {
    "type": "SPACE",
    "string": " "
  },
  {
    "type": "WORD",
    "string": "su",
    "mi": "Sgi",
    "lemma": "su"
  },
  {
    "type": "SPACE",
    "string": " "
  },
  {
    "type": "WORD",
    "string": "kasa",
    "mi": "Ncfsin-",
    "lemma": "kasa"
  },
  {
    "type": "SPACE",
    "string": " "
  },
  {
    "type": "WORD",
    "string": "kasa",
    "mi": "Vgmp3---n--ni-",
    "lemma": "kasti"
  },
...
]
```

Info about the values of `mi` property can be found here [http://corpus.vdu.lt/en/morph](http://corpus.vdu.lt/en/morph). The set of possible values for the `type` field is `SPACE, SEPARATOR, SENTENCE_END, NUMBER, WORD`.

---
### Author

**Airenas Vaičiūnas**
 
* [github.com/airenas](https://github.com/airenas/)
* [linkedin.com/in/airenas](https://www.linkedin.com/in/airenas/)


---
### License

Copyright © 2021, [Airenas Vaičiūnas](https://github.com/airenas).
Released under the [The 3-Clause BSD License](LICENSE).

---

