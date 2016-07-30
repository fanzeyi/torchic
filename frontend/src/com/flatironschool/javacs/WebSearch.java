package com.flatironschool.javacs;

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

import java.util.concurrent.Callable;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.Future;

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
	private static Integer TOTALDOCUMENTS;//can i take this out?
	protected static JedisIndex index = new JedisIndex(new Jedis("localhost", 6379));
	/**
	* Constructor.
	*
	* @param map
	*/

	public WebSearch(Map<String, Integer> map) {
		//this.map = map;
		this.map = new MapTF(map, term);
		this.n_i = this.map.getTotalFrequency();
		this.mapBM = new MapBM<String, Double>(map, index, term, 1);
	}
	public WebSearch(Map<String, Integer> map, String term, Integer termWeight)
	{
		//this.map = map;
		this.map = new MapTF(map, term);
		this.n_i = this.map.getTotalFrequency();
		this.term = term;
		//this.mapBM = bm25(map, qf_i);
		this.mapBM = new MapBM<String, Double>(map, index, term, termWeight);
	}
	public static MapBM<String, Double> multiSearch(WebSearch t1, WebSearch t2)
	{
		MapBM<String, Double> refined = new MapBM<String, Double>();
		MapBM<String, Double> m1 = t1.mapBM;
		Map<String, Double> m2 = t2.mapBM;
		Set<String> m1_keys = (Set<String>)m1.keySet();
		Set<String> m2_keys = (Set<String>)m2.keySet();
		for(String s:m1_keys)
		{
			if(m2.containsKey(s))
			{
				Double rel = m1.get(s)+m2.get(s);
				refined.put(s,rel);
				m2_keys.remove(s);
			}
			else
			{
				Double rel = m1.get(s);
				refined.put(s, rel);
			}
			//m1_keys.remove(s);
		}
		System.out.println("Empty"+m2_keys.isEmpty());
		for(String s:m2_keys)
		{
			Double rel = m2.get(s);
			refined.put(s, rel);
			//m2_keys.remove(s);
		}
		return refined;
	}


	/**
	* Prints the contents in order of term frequency.
	*
	* @param map

	private void print() {
		List<Entry<String, Integer>> entries = this.mapBM.entrySet().sort();
		for (Entry<String, Integer> entry: entries) {
			System.out.println(entry);
		}
	}
	*/


	/**
	* Performs a search and makes a WebSearch object.
	*
	* @param term
	* @param index
	* @return

	public static WebSearch singleSearch(String term, JedisIndex index, Integer qf_i) {
		Map<String, Integer> map = index.getCounts(term);//mapping from document to term frequency
		Integer n_i = new Integer(index.returnTotal(map));//should be: number of times term i appears in all documents
		System.out.println(term+": n_i: "+n_i);
		TOTALDOCUMENTS = index.numberOfTermCounters();
		return new WebSearch(map, n_i, qf_i);
	}
	*/
	public static WebSearch singleSearch(Map.Entry<String, Integer> query) {
		String term = query.getKey();
		Integer termWeight = query.getValue();

		Map<String, Integer> map = index.getCounts(term);//mapping from document to term frequency
		//Integer totalTermFrequency = new Integer(index.returnTotal(map));//should be: number of times term i appears in all documents
		//Integer totalTermFrequency = new Integer(map.totalFrequency(term));
		//System.out.println(term+": n_i: "+n_i);
		TOTALDOCUMENTS = index.getTotalDocuments();//is this correct?
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

	public static void main(String[] args) throws IOException {
		Scanner in = new Scanner(System.in);
		System.out.println("Please enter search terms");
		String s = in.nextLine();
		UserQuery query = new UserQuery(s);
		/**
		make a JedisIndex
		Jedis jedis = JedisMaker.make();
		JedisIndex index = new JedisIndex(jedis);
		*/
		List<WebSearch> searchResults = search(query);
		System.out.println("Multi print");
		MapBM<String, Double> multi = multiSearch(searchResults.get(0), searchResults.get(1));
		multi.print();
		/*Set<String> keySet = query.keySet();
		//System.out.println(arrString(arr));
		//JedisIndex.loadIndex(index);
		//processRelevance(keySet);
		// search for the first term

		List<WebSearch> list = new LinkedList<WebSearch>();
		for(String t:keySet)
		{
			String term = t;
			Integer qf_i = query.get(t);
			WebSearch search = singleSearch(term, index, qf_i);
			search.print();
			list.add(search);
		}*/
		System.out.println("end");
	}
}
