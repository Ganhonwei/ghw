package com.ghw.event.inter;

public interface IEventPost {
    interface Test extends IEventPost {
        void test(int i);
    }

    interface Student extends IEventPost {
        void test(int i);
    }
}
