package codeu.unnamed.frontend;
import java.io.IOException;
import java.util.*;
import java.util.Map.Entry;
import java.util.stream.Collectors;

import org.springframework.stereotype.Component;
import redis.clients.jedis.Jedis;
import redis.clients.jedis.Tuple;

/**
* Represents the results of a search query.
*
*/
@Component
public class WebSearch {
	// map from document containing term t to BM25 score
	private MapBM<Tuple> mapBM;
	//number of times term t appears in all documents of collection containing term t
	private String term;
	private static Integer totalDocuments;
	protected static JedisIndex index = new JedisIndex(new Jedis("localhost", 6379));

    /**
	* Constructor.
	*
     * @param map
     */
	public WebSearch(Set<Tuple> map, String term, Integer termWeight)
	{
		this.term = term;
		this.mapBM = new MapBM<Tuple>(map, index, term, termWeight);
	}
	public WebSearch()
	{
		this.term = null;
		this.mapBM = null;
	}

	public static ResultMap multiSearch(List<WebSearch> list)
	{
		List<Map<String, Double>> bmList = new ArrayList<>();
		List<String> terms = new ArrayList<>();

		for(WebSearch search: list) {
			terms.add(search.term);

			bmList.add(search.mapBM.getMap());
		}

		Map<String, Double> resMap = one(bmList);

		String[] queries = terms.toArray(new String[0]);

		return new ResultMap(resMap, queries);
	}
	private static Map<String, Double> one(List<Map<String, Double>> maps)
	{
		if(maps.size() == 0)
		{
			return new HashMap<>();
		}
		else if(maps.size()==1)
		{
			return maps.get(0);
		}
		else
		{
			Map<String, Double> l1 = maps.get(0);
			Map<String, Double> l2 = maps.get(1);
			List<Map<String, Double>> l3 = maps.subList(2, maps.size());
			return two(two(l1,l2),one(l3));
		}
	}
	private static Map<String, Double> two(Map<String, Double> m1, Map<String, Double> m2)
	{
		Map<String, Double> combined = new HashMap<>();
		Set<String> m1_keys = m1.keySet();
		Set<String> m2_keys = m2.keySet();
		for(String s:m1_keys)
		{
			if(m2.containsKey(s))
			{
				Double rel = m1.get(s)+m2.get(s);
				combined.put(s,rel);
				m2_keys.remove(s);
			}
			else
			{
				Double rel = m1.get(s);
				combined.put(s, rel);
			}
		}
		for(String s:m2_keys)
		{
			Double rel = m2.get(s);
			combined.put(s, rel);
		}
		return combined;
	}

	/**
	* Performs a search and makes a WebSearch object.
	*
	* @param term
	* @param index
	* @return
	*/
	public static WebSearch singleSearch(Map.Entry<String, Integer> query) {
		String term = query.getKey();
		Integer termWeight = query.getValue();

		Set<Tuple> map = index.getCounts(term); // mapping from document to term frequency
//		totalDocuments = index.getTotalDocuments();

		return new WebSearch(map, term, termWeight);

	}
	/**
	* Returns a list of WebSearch objects from user query - each WebSearch objects
	* corresponds to a search term
	* TO DO: Implement Java Future
	*/
	public static List<WebSearch> search(UserQuery query)
	{
		return query.getQueries().stream().map(WebSearch::singleSearch).collect(Collectors.toList());
	}
	
   public List<String> processQueries(String[] q)
   {
	   UserQuery query = new UserQuery(q);
	   List<WebSearch> searchResults = search(query);
	   ResultMap multi = multiSearch(searchResults);
	   return multi.returnResultSet();
   }
	
	public static void main(String[] args) throws IOException {
		String[] q = {"java", "project"};
		System.out.println(new WebSearch().processQueries(q));
	}
}

