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

public class MapBM<K,V> extends HashMap<K,V> implements MapRelevance<K,V>
{
	protected Map<String,Double> map;
	protected JedisIndex index;
	protected String term;
	private Integer termWeight;
	private final double K1 = 1.2;
	private final double K2 = 100;
	
	public MapBM(Map<String, Integer> old, JedisIndex index, String term, Integer termWeight)
	{
		this.index = index;
		this.term = term;
		this.termWeight = termWeight;
		this.map = convert(old);
	}
	public Map<String, Double> getMap()
	{
		return this.map;
	}
	public Double getRel(String s)
	{
		return this.map.get(s) == null ? 0.0 : this.map.get(s);
	}
	public int size()
	{
		return this.map.size();
	}
	public Set<K> keySet()
	{
		return (Set<K>) this.map.keySet();
	}
	/**
	 * Takes in a mapping from documents to term frequency and returns a mapping
	 * from documents to BM25 relevance score
	 */
	public Map<String, Double> convert(Map<String, Integer> old)
	{
		Map<String, Double> newMap = new HashMap<String, Double>();
		Set keySet = old.keySet();
		for(Object s: keySet.toArray())
		{
			String url = s.toString();
			newMap.put(url, getSingleRelevance(url));
		}
		return newMap;
	}
	public Double getSingleRelevance(String url)
	{
		try {
			Double averageDocLength = index.getAverageDocLength();
			Integer docLength = index.termsIndexedOnPage(url);
			Integer termFrequency = index.getCount(url, this.term);
			Double k = new Double(1.2*(0.25+(0.75*(docLength/averageDocLength))));
			Double a = (0.5/0.5);
			Integer numberOfDocsContainingTerm = index.numberOfDocsContainingTerm(term);
			Integer totalDocuments = index.getTotalDocuments();
			Double b = ((numberOfDocsContainingTerm+0.5)/(totalDocuments-numberOfDocsContainingTerm+0.5));
			Double c = ((1.2+1)*(termFrequency))/(k+termFrequency);
			Double d = (double)(((100+1)*(termWeight))/(100+termWeight));//
			Double result = ((Math.log(a/b))*c*d)*100;
			//System.out.println(result);
			return result;
		}
		catch (Exception e)
		{
			System.out.println("Message:"+e.getMessage());
		}
		return null;
	}

	public List<Entry<String, Double>> sort() {
		Set<Entry<String,Double>> entrySet = this.map.entrySet();
		List<Entry<String, Double>> list = new LinkedList<Entry<String, Double>>(entrySet);
		System.out.println("showlist");
		showList(list);
		System.out.println("showlist");

		Comparator<Entry<String, Double>> c = new Comparator<Entry<String, Double>>()
		{
			@Override
			public int compare(Entry<String, Double> e1, Entry<String, Double> e2)
			{
				Double val_e1 = new Double(e1.getValue().toString());
				Double val_e2 = new Double(e2.getValue().toString());
				return val_e1.compareTo(val_e2);
			}
		};
		Collections.sort(list, c);
		return list;
	}
	public void showList(List<Entry<String,Double>> list)
	{
		System.out.println("Size:"+list.size());
		for(Entry<String, Double> e: list)
		{
			System.out.println("Entry:"+e);
		}
	}
	public void entryView()
	{
		List<Entry<String, Double>> entries = new LinkedList<Entry<String, Double>>(this.map.entrySet());
		System.out.println(entries.toString());
	}

	public void print() {
		List<Entry<String, Double>> entries = sort();
		for (Entry<String, Double> entry: entries) {
			System.out.println(entry);
		}
	}
	
	
}
