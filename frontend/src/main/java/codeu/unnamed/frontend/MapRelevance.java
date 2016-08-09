package codeu.unnamed.frontend;
import java.util.List;
import java.util.Map;
import java.util.Map.Entry;

public interface MapRelevance<K,V>
{
  /**
  * Takes in a mapping from documents to term frequency and returns a mapping
  * from documents to BM25 relevance score
  */
  Map<String,Double> convert(Map<String,Integer> map);
  /**
  * Calculates the relevance score for each document within a collection
  */
  Double getSingleRelevance(String key);
  /**
  * Sorts the map in ascending order
  */
  List<Entry<String, Double>> sort();
  /**
  * Prints the entries of the given map
  */
  void print();

}
