//
//  GBAppDelegate.h
//  Gopherbox
//
//  Created by Steven Schmatz on 8/3/14.
//  Copyright (c) 2014 Gopherbox. All rights reserved.
//

#import <Cocoa/Cocoa.h>

@interface GBAppDelegate : NSObject <NSApplicationDelegate>

@property (weak) IBOutlet NSMenu *statusMenu;
@property (strong, nonatomic) NSStatusItem *statusBar;

- (void)loginVerified:(BOOL)verified withUsername:(NSString *)usernameGiven andPassword:(NSString *)passwordGiven;
- (NSString *)runBinaryWithName:(NSString *)binaryName andArguments:(NSArray *)arguments;

@end