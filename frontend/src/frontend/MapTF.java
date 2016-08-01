/**
 * 
 */
package frontend;


import java.util.Collections;
import java.util.Comparator;
import java.util.HashMap;
import java.util.LinkedList;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.Map.Entry;

import redis.clients.jedis.Jedis;
/**
 * @author Medea
 *
 */
public class MapTF {
	private Map<String, Integer> map;
	private String term;//term associated with this map
	/**
	 * Constructor
	 */
	public MapTF(Map<String, Integer> map, String term)
	{
		this.map = map;
		this.term = term;
	}
	public void print()
	{
		List<Entry<String, Integer>> entries = this.sort();
		for (Entry<String, Integer> entry: entries) {
			System.out.println(entry);
		}
	}
	/**
	 * Sort the results by relevance.
	 *
	 * @return List of entries with URL and relevance.
	 */
	public List<Entry<String, Integer>> sort() {
		Set entrySet = this.map.entrySet();
		List list = new LinkedList<Entry<String, Integer>>(entrySet);
		Comparator<Entry<String, Integer>> c = new Comparator<Entry<String, Integer>>()
		{
			@Override
			public int compare(Entry<String, Integer> e1, Entry<String, Integer> e2)
			{
				Integer val_e1 = e1.getValue();
				Integer val_e2 = e2.getValue();
				if(val_e1 == val_e2)
				{
					return 0;
				}
				else if(val_e1 > val_e2)
				{
					return 1;
				}
				return -1;
			}
		};
		Collections.sort(list, c);
		return list;
	}
	/**
	 * Looks up the relevance of a given URL.
	 *
	 * @param url
	 * @return
	 */
	public Integer getRelevance(String url) {
		Integer relevance = map.get(url);
		return relevance==null ? 0: relevance;
	}
	public Integer getTotalFrequency()
	{
		int count = 0;
		List<Integer> values = new LinkedList(map.values());
		for(Integer n: values)
		{
			count+=n;
		}
		return new Integer(count);
	}
}

