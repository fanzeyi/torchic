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

public class MapBM<K,V> extends HashMap<K,V> implements MapRelevance<K,V>
{
  protected Map<String,Double> map;

  public MapBM()
  {
    this.map = new HashMap<String,Double>();
  }
  public MapBM(Map<String, Integer> map)
  {
    this.map = convert(map);
  }
  /**
  * Takes in a mapping from documents to term frequency and returns a mapping
  * from documents to BM25 relevance score
  */
  public Map<String, Double> convert(Map<String, Integer> old, Integer qf_i)//can i take out the qf_i field?
	{
		Map<String, Double> newMap = new HashMap<String, Double>();
		Set keySet = old.keySet();
		for(Object s: keySet.toArray())
		{
			newMap.put(s.toString(), getSingleRelevance(s.toString(), qf_i));
		}
		return newMap;
	}
  public Double getSingleRelevance(String url, Integer qf_i)
	{
		Integer f_i = getRelevance(url);//termcounter - number of times the word appears
		Double k = 1.2;//fix - need to include doc length param
		Double a = (0.5/0.5);
		Double b = ((this.n_i+0.5)/((TOTALDOCUMENTS*500)-this.n_i+0.5));//weighted TOTALDOCUMENT because of sample size
		Double c = ((1.2+1)*(f_i))/(k+f_i);
		Double d = (double)(((100+1)*(qf_i))/(100+qf_i));//
		Double result = (Math.log(a/b))*c*d*100;

		return result;
	}
  public List<Entry<String, Double>> sort() {
    Set entrySet = this.map.entrySet();
		List list = new LinkedList<Entry<String, Double>>(entrySet);
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

  protected void print() {
		List<Entry<String, Double>> entries = sort();
		for (Entry<String, Double> entry: entries) {
			System.out.println(entry);
		}
	}
}
