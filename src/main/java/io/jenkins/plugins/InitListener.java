package io.jenkins.plugins;

import hudson.init.InitReactorListener;
import org.jvnet.hudson.reactor.Milestone;
import org.jvnet.hudson.reactor.Task;
import org.kohsuke.MetaInfServices;

/**
 * @author <a href="mailto:nicolas.deloof@gmail.com">Nicolas De Loof</a>
 */
// @MetaInfServices doesn't work :'(
public class InitListener implements InitReactorListener {

    @Override
    public void onTaskStarted(Task task) {
        System.out.println("started :" + task.toString());
    }

    @Override
    public void onTaskCompleted(Task task) {
        System.out.println("completed :" + task.toString());
    }

    @Override
    public void onTaskFailed(Task task, Throwable throwable, boolean b) {
        System.err.println("failed :" + task.toString());
        System.exit(127);
    }

    @Override
    public void onAttained(Milestone milestone) {
        System.out.println("attained :" + milestone.toString());
    }
}
