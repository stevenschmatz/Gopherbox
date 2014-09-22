//
//  GBClientThreadObject.h
//  Gopherbox
//
//  Created by Steven Schmatz on 8/8/14.
//  Copyright (c) 2014 Gopherbox. All rights reserved.
//

#import <Foundation/Foundation.h>

@interface GBClientThreadObject : NSObject

@property (nonatomic, strong) NSThread *clientThread;

- (void)startGopherboxThreadWithUsername:(NSString *)username andPassword:(NSString *)password ;

@property (nonatomic, strong) NSString *username;
@property (nonatomic, strong) NSString *password;

- (void)startGopherbox;
- (void)stopGopherbox;

@end
