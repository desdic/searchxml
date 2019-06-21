# Searchxml

This is a small tool to search though XML files via namespace, attributes, tags and content

## Usage

```sh
Usage of bin/searchxml:
  -attr value
        Attr key/value
  -color
        Enable color highlight
  -content string
        Regex to match content
  -cores int
        Number of cores to run on (default 4)
  -namespace string
        Regex to match a namespace
  -tag string
        Regex to match a tag
```

## Example
```sh
# find ~/xbrls -iname '*.xml' -type f|xargs ./xmlparse -tag TypeOfAuditorAssistance -content '(?i)ingen bistand' -color
/home/kgn/xbrls/nonfsa/1149374.xml
Namespace: http://xbrl.dcca.dk/cmn
Tag: TypeOfAuditorAssistance
Attributes: [{{ contextRef} duration_Q3_2016_only}]
Content:
Ingen bistand


/home/kgn/xbrls/nonfsa/1149174.xml
Namespace: http://xbrl.dcca.dk/cmn
Tag: TypeOfAuditorAssistance
Attributes: [{{ contextRef} duration_Q3_2016_only}]
Content:
Ingen bistand


/home/kgn/xbrls/nonfsa/1150256.xml
Namespace: http://xbrl.dcca.dk/cmn
Tag: TypeOfAuditorAssistance
Attributes: [{{ contextRef} duration_Q3_2016_only}]
Content:
Ingen bistand
....
```
