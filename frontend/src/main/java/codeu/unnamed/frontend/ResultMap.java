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

	public ResultMap(Map<String, Double> map) {
		this.map = map;
	}
	public List<String> returnResultSet()
	{
		List<String> entries = sort();

		return entries;
	}

	public List<String> sort() {
		List<String> docIds = new LinkedList<>(map.keySet());
		Comparator<String> comp = (o1, o2) -> Double.compare(map.get(o2), map.get(o1));
		Collections.sort(docIds, comp);
		return docIds;
	}

}
