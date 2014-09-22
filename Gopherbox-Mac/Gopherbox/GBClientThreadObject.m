//
//  GBClientThreadObject.m
//  Gopherbox
//
//  Created by Steven Schmatz on 8/8/14.
//  Copyright (c) 2014 Gopherbox. All rights reserved.
//

#import "GBClientThreadObject.h"
#import "GBAppDelegate.h"

@implementation GBClientThreadObject {
    NSPipe *inputPipe;
}

- (void)startGopherboxThreadWithUsername:(NSString *)username andPassword:(NSString *)password {
    self.username = username;
    self.password = password;
    self.clientThread = [[NSThread alloc] initWithTarget:self
                                               selector:@selector(startGopherbox)
                                                 object:nil];
    [self.clientThread start];  // Actually create the thread
}

- (void)startGopherbox {
    NSString *fullPath = [[NSUserDefaults standardUserDefaults] valueForKey:@"gopherboxFolderPath"];
    NSString *path = [fullPath substringFromIndex:7];
    NSArray *arguments = [[NSArray alloc] initWithObjects:path, self.username, self.password, nil];
    [self runBinaryWithName:@"Client" andArguments:arguments];
}

// runBinaryWithName runs the binary with the given name and arguments in the resources folder.
- (NSString *)runBinaryWithName:(NSString *)binaryName andArguments:(NSArray *)arguments {
    
    NSString *launchPath = [[NSBundle mainBundle] pathForResource:binaryName ofType:nil];
    
    NSTask *task;
    task = [[NSTask alloc] init];
    [task setLaunchPath: launchPath];
    
    [task setArguments: arguments];
    
    NSPipe *pipe;
    pipe = [NSPipe pipe];
    [task setStandardOutput: pipe];
    
    inputPipe = [NSPipe pipe];
    [task setStandardInput: inputPipe];
    
    NSFileHandle *file;
    file = [pipe fileHandleForReading];
    
    [task launch];
    
    NSData *data;
    data = [file readDataToEndOfFile];
    
    NSString *results;
    results = [[NSString alloc] initWithData: data encoding: NSUTF8StringEncoding];
    
    return results;
}

- (void)stopGopherbox {
    NSFileHandle *file = [inputPipe fileHandleForWriting];
    NSData *newData = [@"kill\n" dataUsingEncoding:NSUTF8StringEncoding];
    [file writeData:newData];
}

@end
