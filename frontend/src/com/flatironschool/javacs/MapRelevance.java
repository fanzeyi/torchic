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

public interface MapRelevance<K,V>
{
  /**
  * Takes in a mapping from documents to term frequency and returns a mapping
  * from documents to BM25 relevance score
  */
  Map<K,V> convert(Map<K,T> map, Integer i);
  /**
  * Calculates the relevance score for each document within a collection
  */
  V getSingleRelevance(K key, Integer i);
  /**
  * Sorts the map in ascending order
  */
  List<Entry<K, V>> sort();
  /**
  * Prints the entries of the given map
  */
  void print(Map<K,V> map);

}
