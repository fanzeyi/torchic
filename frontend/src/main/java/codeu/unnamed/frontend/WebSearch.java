package codeu.unnamed.frontend;
import java.io.IOException;
import java.util.*;
import java.util.Map.Entry;
import java.util.stream.Collectors;

import org.springframework.stereotype.Component;
import redis.clients.jedis.Jedis;
import redis.clients.jedis.Tuple;

/**
* Represents the results of a search query.
*
*/
@Component
public class WebSearch {
	// map from document containing term t to BM25 score
	private MapBM<Tuple> mapBM;
	//number of times term t appears in all documents of collection containing term t
	private String term;
	protected static JedisIndex index = new JedisIndex(new Jedis("localhost", 6379));

    /**
	* Constructor.
	*
     * @param map
     */
	public WebSearch(Set<Tuple> map, String term, Integer termWeight)
	{
		this.term = term;
		this.mapBM = new MapBM<Tuple>(map, index, term, termWeight);
	}
	public WebSearch()
	{
		this.term = null;
		this.mapBM = null;
	}

	public static ResultMap multiSearch(List<WebSearch> list)
	{
		List<Map<String, Double>> bmList = new ArrayList<>();
		List<String> terms = new ArrayList<>();

		for(WebSearch search: list) {
			terms.add(search.term);

			bmList.add(search.mapBM.getMap());
		}

        Map<String, Double> result = bmList.stream().reduce(new HashMap<>(), (res, elm) -> {
            for (Map.Entry<String, Double> entry : elm.entrySet()) {
                res.merge(entry.getKey(), entry.getValue(), (a, b) -> a+b);
            }
            return res;
        });

		return new ResultMap(result);
	}

	/**
	* Performs a search and makes a WebSearch object.
	*
	* @param term
	* @param index
	* @return
	*/
	public static WebSearch singleSearch(Map.Entry<String, Integer> query) {
		String term = query.getKey();
		Integer termWeight = query.getValue();

		Set<Tuple> map = index.getCounts(term); // mapping from document to term frequency

		return new WebSearch(map, term, termWeight);

	}
	/**
	* Returns a list of WebSearch objects from user query - each WebSearch objects
	* corresponds to a search term
	* TO DO: Implement Java Future
	*/
	public static List<WebSearch> search(UserQuery query)
	{
		return query.getQueries().stream().map(WebSearch::singleSearch).collect(Collectors.toList());
	}
	
   public List<String> processQueries(String[] q)
   {
	   UserQuery query = new UserQuery(q);
	   List<WebSearch> searchResults = search(query);
	   ResultMap multi = multiSearch(searchResults);
	   return multi.returnResultSet();
   }
	
	public static void main(String[] args) throws IOException {
		String[] q = {"java", "project"};
		System.out.println(new WebSearch().processQueries(q));
	}
}

