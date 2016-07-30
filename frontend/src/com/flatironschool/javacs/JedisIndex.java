package com.flatironschool.javacs;

import java.io.IOException;
import java.util.Map;
import java.util.Map.Entry;
import java.util.Set;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;

import org.jsoup.select.Elements;

import redis.clients.jedis.Jedis;
import redis.clients.jedis.Transaction;

/**
 * Represents a Redis-backed web search index.
 *
 */
public class JedisIndex {

	private Jedis jedis;
	private int totalDocsIndexed;
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
		return totalDocsIndexed;
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
	 * Adds a URL to the set associated with `term`.
	 *
	 * @param term
	 * @param tc

	public void add(String term, TermCounter tc) {
		jedis.sadd(urlSetKey(term), tc.getLabel());
	}
	 */

	 /**
 	 * Returns the number of times the given term appears at the given URL.
 	 *
 	 * @param url
 	 * @param term
 	 * @return
 	 */
 	public Integer getCount(String url, String term) {
		String termKey = termURLs(term);
 		int count = jedis.zscore(termKey,url).intValue();
 		return new Integer(count);
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
	public int returnTotal(Map<String, Integer> map)
	{
		int total = 0;
		for(String s:(Set<String>) map.keySet())
		{
			total+=map.get(s);
		}
		return total;
	}
	*/
	/**
	 * Add a page to the index.
	 *
	 * @param url         URL of the page.
	 * @param paragraphs  Collection of elements that should be indexed.

	public void indexPage(String url, Elements paragraphs) {
		System.out.println("Indexing " + url);

		// make a TermCounter and count the terms in the paragraphs
		TermCounter tc = new TermCounter(url);
		tc.processElements(paragraphs);

		// push the contents of the TermCounter to Redis
		pushTermCounterToRedis(tc);
		totalDocsIndexed++;
	}
	*/

	/**
	 * Pushes the contents of the TermCounter to Redis.
	 *
	 * @param tc
	 * @return List of return values from Redis.

	public List<Object> pushTermCounterToRedis(TermCounter tc) {
		Transaction t = jedis.multi();

		String url = tc.getLabel();
		String hashname = termCounterKey(url);

		// if this page has already been indexed; delete the old hash
		t.del(hashname);

		// for each term, add an entry in the termcounter and a new
		// member of the index
		for (String term: tc.keySet()) {
			Integer count = tc.get(term);
			t.hset(hashname, term, count.toString());
			t.sadd(urlSetKey(term), url);
		}
		List<Object> res = t.exec();
		return res;
	}
	*/
	/**
	 * Prints the contents of the index.
	 *
	 * Should be used for development and testing, not production.

	public void printIndex() {
		// loop through the search terms
		for (String term: termSet()) {
			System.out.println(term);

			// for each term, print the pages where it appears
			Set<String> urls = getURLs(term);
			for (String url: urls) {
				Integer count = getCount(url, term);
				System.out.println("    " + url + " " + count);
			}
		}
	}
	 */

	/**
	 * Returns the set of terms that have been indexed.
	 *
	 * Should be used for development and testing, not production.
	 *
	 * @return

	public Set<String> termSet() {
		Set<String> keys = urlSetKeys();
		Set<String> terms = new HashSet<String>();
		for (String key: keys) {
			String[] array = key.split(":");
			if (array.length < 2) {
				terms.add("");
			} else {
				terms.add(array[1]);
			}
		}
		return terms;
	}
	 */

	/**
	 * Returns URLSet keys for the terms that have been indexed.
	 *
	 * Should be used for development and testing, not production.
	 *
	 * @return

	public Set<String> urlSetKeys() {
		return jedis.keys("URLSet:*");
	}
	 */

	/**
	 * Returns TermCounter keys for the URLS that have been indexed.
	 *
	 * Should be used for development and testing, not production.
	 *
	 * @return

	public Set<String> termCounterKeys() {
		return jedis.keys("TermCounter:*");
	}
	public Integer numberOfTermCounters()
	{
		return termCounterKeys().size();
	}
	*/

	/**
	 * Deletes all URLSet objects from the database.
	 *
	 * Should be used for development and testing, not production.
	 *
	 * @return

	public void deleteURLSets() {
		Set<String> keys = urlSetKeys();
		Transaction t = jedis.multi();
		for (String key: keys) {
			t.del(key);
		}
		t.exec();
	}
	*/

	/**
	 * Deletes all URLSet objects from the database.
	 *
	 * Should be used for development and testing, not production.
	 *
	 * @return

	public void deleteTermCounters() {
		Set<String> keys = termCounterKeys();
		Transaction t = jedis.multi();
		for (String key: keys) {
			t.del(key);
		}
		t.exec();
	}
	*/

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
		Jedis jedis = JedisMaker.make();
		JedisIndex index = new JedisIndex(jedis);

		//index.deleteTermCounters();
		//index.deleteURLSets();
		//index.deleteAllKeys();
		//loadIndex(index);
		//System.out.println("#tc:"+index.numberOfTermCounters());


		/*Map<String, Integer> map = index.getCountsFaster("the");
		for (Entry<String, Integer> entry: map.entrySet()) {
			//System.out.println("the");
			System.out.println(entry);
		}*/

	}

	/**
	 * Stores two pages in the index for testing purposes.
	 *
	 * @return
	 * @throws IOException

	public static void loadIndex(JedisIndex index) throws IOException {
		WikiFetcher wf = new WikiFetcher();

		String url = "https://en.wikipedia.org/wiki/Java_(programming_language)";
		Elements paragraphs = wf.fetchWikipedia(url);
		index.indexPage(url, paragraphs);

		url = "https://en.wikipedia.org/wiki/Programming_language";
		paragraphs = wf.readWikipedia(url);
		index.indexPage(url, paragraphs);

		url = "https://en.wikipedia.org/wiki/Mathematics";
		paragraphs = wf.readWikipedia(url);
		index.indexPage(url, paragraphs);

		url = "https://en.wikipedia.org/wiki/Awareness";
		paragraphs = wf.readWikipedia(url);
		index.indexPage(url, paragraphs);

		url = "https://en.wikipedia.org/wiki/Science";
		paragraphs = wf.readWikipedia(url);
		index.indexPage(url, paragraphs);

		url = "https://en.wikipedia.org/wiki/Peloponnesian_War";
		paragraphs = wf.fetchWikipedia(url);
		index.indexPage(url, paragraphs);

		url = "https://en.wikipedia.org/wiki/Plato";
		paragraphs = wf.fetchWikipedia(url);
		index.indexPage(url, paragraphs);

		url = "https://en.wikipedia.org/wiki/Classical_Greece";
		paragraphs = wf.fetchWikipedia(url);
		index.indexPage(url, paragraphs);

		url = "https://en.wikipedia.org/wiki/Oligarchy";
		paragraphs = wf.fetchWikipedia(url);
		index.indexPage(url, paragraphs);

		url = "https://en.wikipedia.org/wiki/Anatolia";
		paragraphs = wf.fetchWikipedia(url);
		index.indexPage(url, paragraphs);
	}
	 */
}
