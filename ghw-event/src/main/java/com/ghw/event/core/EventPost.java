package com.ghw.event.core;

import org.apache.log4j.Logger;

import java.util.HashSet;
import java.util.Set;
import java.util.function.Consumer;

public class EventPost {
    public final Set<Object> __ = new HashSet<>();
    public final Logger logger = Logger.getLogger(EventPost.class);

    public Set<Object> get__() {
        return __;
    }

    public <T> void  doPost(EEventType type, Class<T> t, Consumer<T> consumer) {
        if (type.getInter() != t) {
            logger.error(new IllegalAccessError("class is mismatch"));
            return;
        }
        for (Object o : __) {
            if (type.getInter().isAssignableFrom(o.getClass())) {
                consumer.accept((T) o);
            }
        }
    }

    public static EventPost getInstance() {
        return Instance.INSTANCE.getEvent();
    }

    public enum Instance {
        INSTANCE,
        ;

        private final EventPost event = new EventPost();

        public EventPost getEvent() {
            return event;
        }
    }
}
