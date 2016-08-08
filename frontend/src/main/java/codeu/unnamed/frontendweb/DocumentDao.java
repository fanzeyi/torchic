package codeu.unnamed.frontendweb;

import org.springframework.data.repository.CrudRepository;

import javax.transaction.Transactional;

/**
 * Created by fanzeyi on 8/4/16.
 */
@Transactional
public interface DocumentDao extends CrudRepository<Document, Long> {

    public Document findByHash(String hash);

    public Document findById(long id);

}
