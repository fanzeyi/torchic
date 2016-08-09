package codeu.unnamed.frontend;

import java.util.Arrays;
import java.util.Collections;
import java.util.Comparator;
import java.util.HashMap;
import java.util.LinkedList;
import java.util.List;
import java.util.Map;
import java.util.Map.Entry;
import java.util.Set;

public class ResultMap {
	
	private Map<String, Double> map;
	private String[] query;
	
	public ResultMap(Map<String, Double> map, String[] query) {
		this.map = map;
		this.query = query;
	}
	public void print()
	{
		List<Entry<String, Double>> entries = sort();
		entries = getNResults(entries, 10);
		for (Entry<String, Double> entry: entries) {
			System.out.println(entry);
		}
	}
	public List<Entry<String, Double>> sort() {
		Set<Entry<String,Double>> entrySet = this.map.entrySet();
		List<Entry<String, Double>> list = new LinkedList<Entry<String, Double>>(entrySet);
		Comparator<Entry<String, Double>> c = new Comparator<Entry<String, Double>>()
		{
			@Override
			public int compare(Entry<String, Double> e1, Entry<String, Double> e2)
			{
				Double val_e1 = e1.getValue();
				Double val_e2 = e2.getValue();
				int cmp = val_e1.compareTo(val_e2);
				return cmp;
			}
		};
		Collections.sort(list, c);
		return list;
	}
	public List<Entry<String, Double>> getNResults(List<Entry<String, Double>> results, int n)
	{
		int count = results.size() > n ? n : results.size();
		return results.subList(0, count);
	}

}
