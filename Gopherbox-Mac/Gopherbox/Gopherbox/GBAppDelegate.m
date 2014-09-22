//
//  GBAppDelegate.m
//  Gopherbox
//
//  Created by Steven Schmatz on 8/3/14.
//  Copyright (c) 2014 Gopherbox. All rights reserved.
//

#import "GBAppDelegate.h"
#import "GBLoginWindowController.h"
#import "GBPreferencesWindowController.h"
#import "GBClientThreadObject.h"

@implementation GBAppDelegate {
    GBLoginWindowController *loginController;
    GBPreferencesWindowController *preferencesController;
    
    BOOL userIsVerified;
    NSString *username;
    NSString *password;
    NSThread *myThread;
    
    GBClientThreadObject *gopherboxThreadObj;
}

// applicationDidFinishLaunching was overridden to init the login window controller.
- (void)applicationDidFinishLaunching:(NSNotification *)aNotification
{
    loginController = [[GBLoginWindowController alloc] initWithWindowNibName:@"GBLoginWindowController"];
    [loginController showWindow:nil];
    [loginController.window makeKeyAndOrderFront:nil];
    
}

// loginVerified is called by GBLoginWindowController to sync login status.
- (void)loginVerified:(BOOL)verified withUsername:(NSString *)usernameGiven andPassword:(NSString *)passwordGiven
{
    userIsVerified = loginController.verified;
    username = loginController.username;
    password = loginController.password;
    
    if (userIsVerified) {
        self.statusBar = [[NSStatusBar systemStatusBar] statusItemWithLength:NSVariableStatusItemLength];
        
        self.statusBar.image = [NSImage imageNamed:@"gopherbox-menubar"];
        self.statusBar.title = @" Gopherbox";
        
        self.statusBar.menu = self.statusMenu;
        self.statusBar.highlightMode = YES;
        
        gopherboxThreadObj = [[GBClientThreadObject alloc] init];
        [gopherboxThreadObj startGopherboxThreadWithUsername:username andPassword:password];
    }
    
    NSLog(@"App delegate got user credentials:\nUsername:%@\nPassword:%@\nVerified:%hhd", username, password, userIsVerified);
}

// showPreferences displays the preferences window.
- (IBAction)showPreferences:(id)sender {
    preferencesController = [[GBPreferencesWindowController alloc] initWithWindowNibName:@"GBPreferencesWindowController"];
    [preferencesController showWindow:nil];
    [preferencesController.window makeKeyAndOrderFront:nil];
}

// showGopherboxFolder opens the Gopherbox folder in Finder.
- (IBAction)showGopherboxFolder:(id)sender {
    [[NSWorkspace sharedWorkspace] openURL:[NSURL URLWithString:[[NSUserDefaults standardUserDefaults] valueForKey:@"gopherboxFolderPath"]]];
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
    
    NSFileHandle *file;
    file = [pipe fileHandleForReading];
    
    [task launch];
    
    NSData *data;
    data = [file readDataToEndOfFile];
    
    NSString *results;
    results = [[NSString alloc] initWithData: data encoding: NSUTF8StringEncoding];
    
    return results;
}

- (void)applicationWillTerminate:(NSNotification *)notification {
    NSLog(@"Application terminating!");
    [gopherboxThreadObj stopGopherbox];
}

@end
