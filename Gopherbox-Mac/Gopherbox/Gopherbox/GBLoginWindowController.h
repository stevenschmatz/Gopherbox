//
//  GBLoginWindowController.h
//  Gopherbox
//
//  Created by Steven Schmatz on 8/3/14.
//  Copyright (c) 2014 Gopherbox. All rights reserved.
//

#import <Cocoa/Cocoa.h>

@interface GBLoginWindowController : NSWindowController

@property (strong) IBOutlet NSWindow *loginWindow;
@property (strong) NSString *username;
@property (strong) NSString *password;
@property bool verified;

@end
