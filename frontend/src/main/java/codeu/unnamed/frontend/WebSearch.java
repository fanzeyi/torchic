package codeu.unnamed.frontend;
import java.io.IOException;
import java.util.Collections;
import java.util.Comparator;
import java.util.HashMap;
import java.util.LinkedList;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.Map.Entry;
import java.util.HashSet;
import java.util.Scanner;

import redis.clients.jedis.Jedis;

/**
* Represents the results of a search query.
*
*/
public class WebSearch {
	// map from document that contains term t to term frequency
	private MapTF map;
	// map from document containing term t to BM25 score
	private MapBM<String, Double> mapBM;
	//number of times term t appears in all documents of collection containing term t
	private Integer n_i;
	private String term;
	private static Integer TOTALDOCUMENTS;
	protected static JedisIndex index = new JedisIndex(new Jedis("localhost", 6379));
	/**
	* Constructor.
	*
	* @param map
	*/

	public WebSearch(Map<String, Integer> map, String term, Integer termWeight)
	{
		this.map = new MapTF(map, term);
		this.n_i = this.map.getTotalFrequency();
		this.term = term;
		this.mapBM = new MapBM<String, Double>(map, index, term, termWeight);
	}
	public WebSearch()
	{
		this.map = null;
		this.n_i = null;
		this.term = null;
		this.mapBM = null;
	}

	public static ResultMap multiSearch(List<WebSearch> list)
	{
		List<Map<String, Double>> bmList = new LinkedList<Map<String, Double>>();
		List<String> terms = new LinkedList<String>();
		for(WebSearch search: list)
		{
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
			return new HashMap<String, Double>();
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
		Map<String, Double> combined = new HashMap<String, Double>();
		Set<String> m1_keys = (Set<String>)m1.keySet();
		Set<String> m2_keys = (Set<String>)m2.keySet();
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

		Map<String, Integer> map = index.getCounts(term);//mapping from document to term frequency
		//Integer totalTermFrequency = new Integer(index.returnTotal(map));//should be: number of times term i appears in all documents
		//Integer totalTermFrequency = new Integer(map.totalFrequency(term));
		//System.out.println(term+": n_i: "+n_i);
		TOTALDOCUMENTS = index.getTotalDocuments();
		//return new WebSearch(map, term, totalTermFrequency, termWeight);
		return new WebSearch(map, term, termWeight);

	}
	/**
	* Returns a list of WebSearch objects from user query - each WebSearch objects
	* corresponds to a search term
	* TO DO: Implement Java Future
	*/
	public static List<WebSearch> search(UserQuery query)
	{
		List<WebSearch> list = new LinkedList<WebSearch>();
		for(Entry<String,Integer> e: query.getQueries())
		{
			list.add(singleSearch(e));
		}
		return list;
	}
	
   public List<String> processQueries(String[] q)
   {
	   UserQuery query = new UserQuery(q);
	   List<WebSearch> searchResults = search(query);
	   ResultMap multi = multiSearch(searchResults);
	   return multi.returnResultSet();
   }
	
	public static void main(String[] args) throws IOException {
		/*Scanner in = new Scanner(System.in);
		System.out.println("Please enter search terms");
		String s = in.nextLine();
		UserQuery query = new UserQuery(s);
		List<WebSearch> searchResults = search(query);
		System.out.println("Multi print");
		//WebSearch one = searchResults.get(0);
		//System.out.println("ONE");
		//one.mapBM.entryView();
		ResultMap multi = multiSearch(searchResults);
		multi.print();
		//List<Entry<String, Double>> results = multi.returnResultSet();
		
		System.out.println("end");*/
		String[] q = {"java", "project"};
		System.out.println(new WebSearch().processQueries(q));
	}
}

