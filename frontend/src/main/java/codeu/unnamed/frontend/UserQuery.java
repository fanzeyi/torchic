package codeu.unnamed.frontend;

import java.util.HashMap;
import java.util.Map;
import java.util.Set;

public class UserQuery
{
	//maps from query to query weight
	private Map<String, Integer> query;
	/**
	 * Constructor
	 */
	public UserQuery(String[] s)
	{
		this.query = processArrayQueries(s);
	}
	/**
	 * Takes query string and creates mapping from query to query weight
	 */
	public Map<String, Integer> processArrayQueries(String[] queries)
	{
		Map<String, Integer> map = new HashMap<>();
		for(String s: queries) {
            map.merge(s, 1, (a, b) -> a + b);
		}
		return map;
	}
	public Set<Map.Entry<String,Integer>> getQueries()
	{
		return this.query.entrySet();
	}
}

