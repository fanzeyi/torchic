package frontend;

import java.io.IOException;
import java.util.Map;
import java.util.Map.Entry;
import java.util.Set;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;

//import org.jsoup.select.Elements;

import redis.clients.jedis.Jedis;
import redis.clients.jedis.Transaction;

/**
 * Represents a Redis-backed web search index.
 *
 */
public class JedisIndex {

	private Jedis jedis;
	private Integer totalDocsIndexed;
	/**
	 * Constructor.
	 *
	 * @param jedis
	 */
	public JedisIndex(Jedis jedis) {
		this.jedis = jedis;
		this.totalDocsIndexed = this.getTotalDocuments();
	}
	public int getTotalDocuments()
	{
		return new Integer(jedis.get("total_documents"));
	}
	//key for a set that stores all the words occuring on the given page
	private String urlSet(String url)
	{
		return "url:"+url;
	}
	//key for number of words on given page
	private String wordCountKey(String url)
	{
		return "count:"+url;
	}
	public Integer getPageWordCount(String url)
	{
		return Integer.valueOf(jedis.get(wordCountKey(url)));
	}
	//key for a sorted set sorting all URLs that contain term t with its ranking at
	//that page
	private String termURLs(String term)
	{
		return "term:"+term;
	}
	/**
	 * Looks up a search term and returns a set top 20 URLs.
	 *
	 * @param term
	 * @return Set of URLs.
	 */
	public Set<String> getURLs(String term) {
		//Set<String> set = jedis.smembers(urlSetKey(term));
		String termKey = termURLs(term);
		Set<String> set = jedis.zrange(termKey,0,20);
		return set;
	}
	//looks up a search term and returns a set of ranked urls from given indices
	public Set<String> getURLs(String term, int start, int end)
	{
		String termKey = termURLs(term);
		Set<String> set = jedis.zrange(termKey, start, end);
		return set;
	}
	public int numberOfDocsContainingTerm(String term)
	{
		String key = termURLs(term);
		return jedis.zcard(key).intValue();
	}
	public Integer totalTermsIndexed()
	{
		return new Integer(jedis.scard("terms").intValue());
	}
	public Integer termsIndexedOnPage(String url)
	{
		String key = urlSet(url);
		return new Integer(jedis.scard(key).toString());
	}
	public Double getAverageDocLength()
	{
		Double result = new Double(this.totalTermsIndexed()/this.totalDocsIndexed);
		return result;
	}
	/**
	 * Checks whether a given URL is indexed
	 *
	 * @param url
	 * @return
	 */
	public boolean isIndexed(String url) {
		String key = urlSet(url);
		return jedis.exists(key);
	}


	 /**
 	 * Returns the number of times the given term appears at the given URL.
 	 *
 	 * @param url
 	 * @param term
 	 * @return
 	 */
 	public Integer getCount(String url, String term) {
		String termKey = termURLs(term);
 		Integer count = new Integer(jedis.zscore(termKey,url).intValue());
 		return count == null ? 0 : count;
 	}

	/**
	 * Looks up a term and returns a map from URL to count.
	 *
	 * @param term
	 * @return Map from URL to count.
	 */
	public Map<String, Integer> getCounts(String term) {
		Map<String, Integer> map = new HashMap<String, Integer>();
		Set<String> urls = getURLs(term);
		for (String url: urls) {
			Integer count = getCount(url, term);
			map.put(url, count);
		}
		return map;
	}

	/**
	 * Looks up a term and returns a map from URL to count.
	 *
	 * @param term
	 * @return Map from URL to count.
	 */
	public Map<String, Integer> getCountsFaster(String term) {
		// convert the set of strings to a list so we get the
		// same traversal order every time
		List<String> urls = new ArrayList<String>();
		urls.addAll(getURLs(term));
		System.out.println("Size: "+urls.size());
		// construct a transaction to perform all lookups
		Transaction t = jedis.multi();
		for (String url: urls) {
			String redisKey = termURLs(term);
			t.hget(redisKey, url);
		}
		List<Object> res = t.exec();

		// iterate the results and make the map
		Map<String, Integer> map = new HashMap<String, Integer>();
		int i = 0;
		for (String url: urls) {
			System.out.println(url);
			Integer count = new Integer((String) res.get(i++));
			map.put(url, count);
		}
		return map;
	}


	/**
	 * Deletes all keys from the database.
	 *
	 * Should be used for development and testing, not production.
	 *
	 * @return
	 */
	public void deleteAllKeys() {
		Set<String> keys = jedis.keys("*");
		Transaction t = jedis.multi();
		for (String key: keys) {
			t.del(key);
		}
		t.exec();
		this.totalDocsIndexed = 0;
	}

	/**
	 * @param args
	 * @throws IOException
	 */
	public static void main(String[] args) throws IOException {
		Jedis jedis = new Jedis("localhost",6379);
		JedisIndex index = new JedisIndex(jedis);
		System.out.println("Server is running: "+jedis.ping());
		System.out.println(index.getCount("https://en.wikipedia.org/wiki/File:Orel_Hershiser_1993_(cropped).jpg", "the"));
		//index.deleteTermCounters();
		//index.deleteURLSets();
		//index.deleteAllKeys();
		//loadIndex(index);

	}


}
