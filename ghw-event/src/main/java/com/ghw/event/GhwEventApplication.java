package com.ghw.event;

import com.ghw.event.core.EEventType;
import com.ghw.event.core.EventPost;
import com.ghw.event.inter.IEventPost;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

@SpringBootApplication
public class GhwEventApplication {

    public static void main(String[] args) {
        SpringApplication application = new SpringApplication(GhwEventApplication.class);
        application.run(args);
        for (int i = 0; i < 10000; ++i) {
            long now = System.currentTimeMillis();
            EventPost.getInstance().doPost(EEventType.TEST, IEventPost.Test.class, t -> t.test(1));
            System.out.println("用时," + (System.currentTimeMillis() - now) + "ms");
        }

//        SpringApplication.run(GhwEventApplication.class, args);
    }
}
