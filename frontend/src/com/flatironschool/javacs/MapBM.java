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

public class MapBM<K,V> extends HashMap<K,V> implements MapRelevance<K,V>
{
  protected Map<String,Double> map;
  protected JedisIndex index;
  protected String term;
  private Integer termWeight;

  public MapBM()
  {
    this.map = new HashMap<String,Double>();
    this.index = new JedisIndex(JedisMaker.make()) ;
  }
  public MapBM(Map<String, Integer> map, JedisIndex index, String term, Integer termWeight)
  {
    this.map = convert(map);
    this.index = index;
    this.term = term;
    this.termWeight = termWeight;
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
			newMap.put(s.toString(), getSingleRelevance(s.toString()));
		}
		return newMap;
	}
  public Double getSingleRelevance(String url)
	{
		Integer termFrequency = index.getCount(url, term);//termcounter - number of times the word appears
		Double k = 1.2;//fix - need to include doc length param
		Double a = (0.5/0.5);
		Double b = ((index.numberOfDocsContainingTerm(term)+0.5)/((index.getTotalDocuments()*500)-index.numberOfDocsContainingTerm(term)+0.5));//weighted TOTALDOCUMENT because of sample size
		Double c = ((1.2+1)*(termFrequency))/(k+termFrequency);
		Double d = (double)(((100+1)*(termWeight))/(100+termWeight));//
		Double result = (Math.log(a/b))*c*d*100;

		return result;
	}

  public List<Entry<String, Double>> sort() {
    Set entrySet = this.map.entrySet();
		List<Entry<String, Double>> list = new LinkedList<Entry<String, Double>>(entrySet);
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

  public void print() {
		List<Entry<String, Double>> entries = sort();
		for (Entry<String, Double> entry: entries) {
			System.out.println(entry);
		}
	}
}
