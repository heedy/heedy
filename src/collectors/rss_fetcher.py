import feedparser

feeds = "http://water.weather.gov/ahps2/rss/obs/lafi3.rss"




d = feedparser.parse('http://feedparser.org/docs/examples/atom10.xml')
'title' in d.feed

'ttl' in d.feed
False
d.feed.get('title', 'No title')
u'Sample feed'

d.feed.get('ttl', 60)
60



<item>
<title>Observation - LAFI3 - Wabash River at Lafayette (Indiana)</title>
<link>http://water.weather.gov/ahps2/hydrograph.php?wfo=ind&#x26;gage=lafi3</link>
<description>Minor Stage: 11 ft&#x3C;br &#x3E;
Minor Flow: 0 kcfs&#x3C;br &#x3E;
&#x3C;br &#x3E;
Latest Observation: 14.06 ft&#x3C;br /&#x3E;
Observation Time: Mar 16, 2015 09:00 AM EDT&#x3C;br /&#x3E;
&#x3C;br /&#x3E;
</description>
<guid isPermaLink="false">http://water.weather.gov/ahps2/hydrograph.php?wfo=ind&#x26;gage=lafi3</guid>
<geo:lat>40.425278</geo:lat>
<geo:long>-86.896389</geo:long>
</item>

