package io.jenkins.plugins;

import hudson.Extension;
import hudson.init.InitMilestone;
import hudson.init.Initializer;
import hudson.model.Result;
import hudson.model.Run;
import hudson.model.TaskListener;
import hudson.model.listeners.RunListener;
import jenkins.model.Jenkins;

import javax.annotation.Nonnull;

/**
 * @author <a href="mailto:nicolas.deloof@gmail.com">Nicolas De Loof</a>
 */
public class JenkinsfileRunner {


    @Initializer(after = InitMilestone.JOB_LOADED)
    public static void run() {
        Jenkins j = Jenkins.getInstance();


    }

    @Extension
    public static RunListener listener = new RunListener() {
        @Override
        public void onCompleted(Run run, @Nonnull TaskListener listener) {
            System.exit(run.getResult() == Result.SUCCESS ? 0 : -1);
        }
    };
}
