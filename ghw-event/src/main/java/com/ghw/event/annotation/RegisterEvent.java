package com.ghw.event.annotation;

import org.springframework.stereotype.Component;

import java.lang.annotation.*;

@Documented
@Target(value = ElementType.TYPE)
@Retention(RetentionPolicy.RUNTIME)
@Inherited
@Component
public @interface RegisterEvent {

}
