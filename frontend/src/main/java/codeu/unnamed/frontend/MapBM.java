package codeu.unnamed.frontend;

import java.util.HashMap;
import java.util.Map;
import java.util.Set;
import java.util.stream.Collectors;

import redis.clients.jedis.Tuple;

public class MapBM<T>
{
    protected Map<String,Double> map;
    protected JedisIndex index;
    protected String term;
    private Integer termWeight;
    private final double K1 = 1.2;
    private final double K2 = 100;

    public MapBM(Set<Tuple> old, JedisIndex index, String term, Integer termWeight)
    {
        this.index = index;
        this.term = term;
        this.termWeight = termWeight;
        this.map = convert(old);
    }
    public Map<String, Double> getMap()
    {
        return this.map;
    }

    /**
     * Takes in a mapping from documents to term frequency and returns a mapping
     * from documents to BM25 relevance score
     */
    public Map<String, Double> convert(Set<Tuple> old) {
        return old.stream().collect(Collectors.toMap(Tuple::getElement, this::getSingleRelevance));
    }

    public double getSingleRelevance(Tuple entry)
    {
        String url = entry.getElement();

        double averageDocLength = index.getAverageDocLength();
        int docLength = index.termsIndexedOnPage(url);
        int termFrequency = (int)entry.getScore();

        double k = K1*(0.25+(0.75*(docLength/averageDocLength)));
        double a = (0.5/0.5);
        int numberOfDocsContainingTerm = index.numberOfDocsContainingTerm(term);
        int totalDocuments = index.getTotalDocuments();
        double b = ((numberOfDocsContainingTerm+0.5)/(totalDocuments-numberOfDocsContainingTerm+0.5));
        double c = ((1.2+1)*(termFrequency))/(k+termFrequency);
        double d = (double)(((100+1)*(termWeight))/(100+termWeight));//
        double result = ((Math.log(a/b))*c*d)*100;

        return result;
    }

}
