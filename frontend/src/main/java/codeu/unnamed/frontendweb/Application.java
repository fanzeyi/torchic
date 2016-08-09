package main.java.codeu.unnamed.frontendweb;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

/**
 * Created by fanzeyi on 8/3/16.
 */
@SpringBootApplication
public class Application {
    @Autowired
    public static void main(String[] args) {
        SpringApplication.run(Application.class, args);
    }
}
