package com.ghw.event.core;

import com.ghw.event.annotation.RegisterEvent;
import org.springframework.beans.BeansException;
import org.springframework.context.ApplicationContext;
import org.springframework.context.ApplicationContextAware;
import org.springframework.stereotype.Component;

import java.util.Map;
import java.util.Set;

@Component
public class AnnotationHandle implements ApplicationContextAware {

    @Override
    public void setApplicationContext(ApplicationContext ctx) throws BeansException {
        Set<Object> set = EventPost.getInstance().get__();
        Map<String, Object> beans = ctx.getBeansWithAnnotation(RegisterEvent.class);
        set.addAll(beans.values());
    }
}
