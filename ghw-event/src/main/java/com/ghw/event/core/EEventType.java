package com.ghw.event.core;

import com.ghw.event.inter.IEventPost;

public enum EEventType {
    UNKNOWN(),
    TEST(IEventPost.Test.class),
    STUDENT(IEventPost.Student.class),
    ;

    private final Class<? extends IEventPost> event;

    EEventType() {
        event = null;
    }

    EEventType(Class<? extends IEventPost> event) {
        this.event = event;
    }

    public Class<? extends IEventPost> getInter() {
        return event;
    }

    public static EEventType getType(Class<?> c) {
        for (EEventType t : values()) {
            if (t.event == c) {
                return t;
            }
        }
        return UNKNOWN;
    }
}
