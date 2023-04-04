package com.ghw.event.explame;

import com.ghw.event.annotation.RegisterEvent;
import com.ghw.event.inter.IEventPost;

@RegisterEvent
public class StudentModel implements IEventPost.Student,IEventPost.Test {
    @Override
    public void test(int i) {
        for (i = 0; i < 1000000; ++i) {
            int j = i << 3;
        }
    }
}
