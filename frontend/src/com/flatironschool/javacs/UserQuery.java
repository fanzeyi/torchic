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

public class UserQuery
{
  //maps from query to query weight
  private Map<String, Integer> query;
  /**
  * Constructor
  */
  public UserQuery(String s)
  {
    this.query = processQueries(s);
  }
  /**
  * Takes query string and creates mapping from query to query weight
  */
	public Map<String,Integer> processQueries(String s)
	{
		Map<String, Integer> map = new HashMap<String, Integer>();
		String[] arr = s.split("\\s");
		for(String k: arr)
		{
			if(map.containsKey(k))
			{
				map.put(s, map.get(k)+1);
			}
			else
			{
				map.put(k,1);
			}
		}
		return map;
	}
  //returns specific query weight
  public Integer getWeight(String s)
  {
    return query.get(s)==null ? 0: query.get(s);
  }
  public Set<Map.Entry<String,Integer>> getQueries()
  {
    return this.query.entrySet();
  }
}
