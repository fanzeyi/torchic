package codeu.unnamed.frontend;

import java.io.IOException;
import java.util.Set;

import redis.clients.jedis.Jedis;
import redis.clients.jedis.Tuple;

/**
 * Represents a Redis-backed web search index.
 *
 */
public class JedisIndex {

	private Jedis jedis;
	/**
	 * Constructor.
	 *
	 * @param jedis
	 */
	public JedisIndex(Jedis jedis) {
		this.jedis = jedis;
	}
	public int getTotalDocuments()
	{
		String total = jedis.get("total_documents");
		if (total == null) return 0;

		return new Integer(total);
	}
	//key for a set that stores all the words occuring on the given page
	private String urlSet(String url)
	{
		return "count:"+url;
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
    public Set<Tuple> getURLs(String term) {
        return getURLs(term, 0, 20);
    }

    public Set<Tuple> getURLs(String term, int start) {
        return getURLs(term, start, start+20);
    }

    //looks up a search term and returns a set of ranked urls from given indices
	public Set<Tuple> getURLs(String term, int start, int end)
	{
		String termKey = termURLs(term);
		Set<Tuple> set = jedis.zrevrangeWithScores(termKey, start, end);
		return set;
	}
	public int numberOfDocsContainingTerm(String term)
	{
		String key = termURLs(term);
		return jedis.zcard(key).intValue();
	}
	public int totalTermsIndexed()
	{
		return new Integer(jedis.get("total_words"));
	}
	public int termsIndexedOnPage(String url)
	{
		String key = urlSet(url);
		return new Integer(jedis.get(key));
	}
	public Double getAverageDocLength()
	{
		int total = this.getTotalDocuments();

		if (total == 0) {
			return 0.0;
		}

		return ((double)totalTermsIndexed()) / ((double)total);
	}

	/**
	 * Looks up a term and returns a map from URL to count.
	 *
	 * @param term
	 * @return Map from URL to count.
	 */
	public Set<Tuple> getCounts(String term) {
		return getURLs(term);
	}

	/**
	 * @param args
	 * @throws IOException
	 */
	public static void main(String[] args) throws IOException {
		Jedis jedis = new Jedis("localhost",6379);
		System.out.println("Server is running: "+jedis.ping());
	}


}
