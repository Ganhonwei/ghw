package com.ghw.event.explame;

import com.ghw.event.annotation.RegisterEvent;
import com.ghw.event.inter.IEventPost;

@RegisterEvent
public class TestModel implements IEventPost.Test {

    @Override
    public void test(int i) {
        System.out.println("sss,test");
    }
}
