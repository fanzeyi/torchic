package frontend;

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
		System.out.println(map.size());
		List<Entry<String, Double>> entries = sort();
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
	public ResultMap combine(ResultMap one, ResultMap two)
	{
		List<String> combinedQueries = Arrays.asList(one.query);
		combinedQueries.addAll(Arrays.asList(two.query));
		String[] queries = combinedQueries.toArray(new String[combinedQueries.size()]);
		HashMap<String, Double> newMap = new HashMap<String, Double>(one.map);
		newMap.putAll(two.map);
		return new ResultMap(newMap, queries);
	}

}
